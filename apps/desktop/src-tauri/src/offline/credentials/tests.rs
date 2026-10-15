use super::*;
use std::{
    collections::HashMap,
    sync::{
        atomic::{AtomicUsize, Ordering},
        mpsc,
    },
    time::Duration,
};

type Key = (String, String);

struct Gate {
    started: Arc<tokio::sync::Notify>,
    release: mpsc::Receiver<()>,
}

#[derive(Default)]
struct FakeCredentials {
    values: Mutex<HashMap<Key, Result<String, CredentialError>>>,
    reads: AtomicUsize,
    writes: AtomicUsize,
    deletes: AtomicUsize,
    gate: Mutex<Option<Gate>>,
    write_error: Mutex<Option<CredentialError>>,
    delete_error: Mutex<Option<CredentialError>>,
}

impl FakeCredentials {
    fn set(&self, environment: &str, account: &str, result: Result<&str, CredentialError>) {
        self.values.lock().unwrap().insert(
            (environment.into(), account.into()),
            result.map(str::to_owned),
        );
    }

    fn block_next_read(&self) -> (Arc<tokio::sync::Notify>, mpsc::Sender<()>) {
        let started = Arc::new(tokio::sync::Notify::new());
        let (release, receiver) = mpsc::channel();
        *self.gate.lock().unwrap() = Some(Gate {
            started: started.clone(),
            release: receiver,
        });
        (started, release)
    }
}

impl CredentialBackend for FakeCredentials {
    fn read(&self, environment: &str, account: &str) -> Result<String, CredentialError> {
        self.reads.fetch_add(1, Ordering::SeqCst);
        let value = self
            .values
            .lock()
            .unwrap()
            .get(&(environment.into(), account.into()))
            .cloned()
            .unwrap_or(Err(CredentialError::Missing));
        let gate = self.gate.lock().unwrap().take();
        if let Some(gate) = gate {
            gate.started.notify_one();
            gate.release.recv_timeout(Duration::from_secs(5)).unwrap();
        }
        value
    }

    fn write(&self, environment: &str, account: &str, token: &str) -> Result<(), CredentialError> {
        self.writes.fetch_add(1, Ordering::SeqCst);
        if let Some(error) = *self.write_error.lock().unwrap() {
            return Err(error);
        }
        self.set(environment, account, Ok(token));
        Ok(())
    }

    fn delete(&self, environment: &str, account: &str) -> Result<(), CredentialError> {
        self.deletes.fetch_add(1, Ordering::SeqCst);
        if let Some(error) = *self.delete_error.lock().unwrap() {
            return Err(error);
        }
        self.values
            .lock()
            .unwrap()
            .remove(&(environment.into(), account.into()));
        Ok(())
    }
}

fn setup() -> (Arc<Credentials>, Arc<FakeCredentials>) {
    let backend = Arc::new(FakeCredentials::default());
    backend.set("development", "account-a", Ok("fixture-session-a"));
    let credentials = Arc::new(Credentials::with_backend(
        "development",
        Some("account-a".into()),
        backend.clone(),
    ));
    (credentials, backend)
}

async fn started(signal: &tokio::sync::Notify) {
    tokio::time::timeout(Duration::from_secs(2), signal.notified())
        .await
        .unwrap();
}

#[tokio::test]
async fn concurrent_background_requests_share_one_os_read_without_blocking_executor() {
    let (credentials, backend) = setup();
    let (signal, release) = backend.block_next_read();
    let mut requests = Vec::new();
    for _ in 0..24 {
        let credentials = credentials.clone();
        requests.push(tokio::spawn(
            async move { credentials.token("account-a").await },
        ));
    }
    started(&signal).await;
    tokio::time::sleep(Duration::from_millis(10)).await;
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
    release.send(()).unwrap();
    for request in requests {
        assert_eq!(request.await.unwrap().unwrap(), "fixture-session-a");
    }
    assert_eq!(
        credentials.token("account-a").await.unwrap(),
        "fixture-session-a"
    );
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn canceled_request_does_not_release_an_open_os_prompt_or_lose_its_result() {
    let (credentials, backend) = setup();
    let (signal, release) = backend.block_next_read();
    let first = {
        let credentials = credentials.clone();
        tokio::spawn(async move { credentials.token("account-a").await })
    };
    started(&signal).await;
    first.abort();
    let second = {
        let credentials = credentials.clone();
        tokio::spawn(async move { credentials.token("account-a").await })
    };
    tokio::time::sleep(Duration::from_millis(10)).await;
    assert!(!second.is_finished());
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
    release.send(()).unwrap();
    assert_eq!(second.await.unwrap().unwrap(), "fixture-session-a");
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn denied_access_is_cached_until_explicit_retry_and_success_is_not_reopened() {
    let (credentials, backend) = setup();
    backend.set(
        "development",
        "account-a",
        Err(CredentialError::Unavailable),
    );
    for _ in 0..10 {
        assert_eq!(
            credentials.token("account-a").await,
            Err(CredentialError::Unavailable)
        );
    }
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
    backend.set("development", "account-a", Ok("fixture-unlocked"));
    assert_eq!(
        credentials.token("account-a").await,
        Err(CredentialError::Unavailable)
    );
    credentials.retry_failed("account-a").unwrap();
    assert_eq!(
        credentials.token("account-a").await.unwrap(),
        "fixture-unlocked"
    );
    credentials.retry_failed("account-a").unwrap();
    assert_eq!(
        credentials.token("account-a").await.unwrap(),
        "fixture-unlocked"
    );
    assert_eq!(backend.reads.load(Ordering::SeqCst), 2);
}

#[tokio::test]
async fn sign_out_invalidates_in_flight_read_and_prevents_old_account_reuse() {
    let (credentials, backend) = setup();
    let (signal, release) = backend.block_next_read();
    let read = {
        let credentials = credentials.clone();
        tokio::spawn(async move { credentials.token("account-a").await })
    };
    started(&signal).await;
    let logout = {
        let credentials = credentials.clone();
        tokio::spawn(async move { credentials.remove("account-a").await })
    };
    tokio::time::timeout(Duration::from_secs(2), async {
        loop {
            if credentials.state.lock().unwrap().account.is_none() {
                break;
            }
            tokio::task::yield_now().await;
        }
    })
    .await
    .unwrap();
    release.send(()).unwrap();
    assert_eq!(read.await.unwrap(), Err(CredentialError::AccountChanged));
    logout.await.unwrap().unwrap();
    assert_eq!(
        credentials.token("account-a").await,
        Err(CredentialError::AccountChanged)
    );
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
    assert_eq!(backend.deletes.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn authentication_replacement_seeds_cache_and_scopes_account_and_environment() {
    let (credentials, backend) = setup();
    assert_eq!(
        credentials.token("account-a").await.unwrap(),
        "fixture-session-a"
    );
    credentials
        .replace("account-a", "fixture-renewed".into())
        .await
        .unwrap();
    assert_eq!(
        credentials.token("account-a").await.unwrap(),
        "fixture-renewed"
    );
    credentials
        .replace("account-b", "fixture-session-b".into())
        .await
        .unwrap();
    assert_eq!(
        credentials.token("account-a").await,
        Err(CredentialError::AccountChanged)
    );
    assert_eq!(
        credentials.retry_failed("account-a"),
        Err(CredentialError::AccountChanged)
    );
    assert_eq!(
        credentials.remove("account-a").await,
        Err(CredentialError::AccountChanged)
    );
    assert_eq!(
        credentials.token("account-b").await.unwrap(),
        "fixture-session-b"
    );
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
    backend.set("production", "account-b", Ok("fixture-production"));
    let production =
        Credentials::with_backend("production", Some("account-b".into()), backend.clone());
    assert_eq!(
        production.token("account-b").await.unwrap(),
        "fixture-production"
    );
    assert_eq!(
        credentials.token("account-b").await.unwrap(),
        "fixture-session-b"
    );
    credentials.remove("account-b").await.unwrap();
    assert_eq!(
        credentials.token("account-b").await,
        Err(CredentialError::AccountChanged)
    );
}

#[tokio::test]
async fn failed_replacement_never_falls_back_to_the_previous_cached_session() {
    let (credentials, backend) = setup();
    credentials.token("account-a").await.unwrap();
    *backend.write_error.lock().unwrap() = Some(CredentialError::Unavailable);
    assert_eq!(
        credentials
            .replace("account-a", "fixture-replacement".into())
            .await,
        Err(CredentialError::Unavailable)
    );
    assert_eq!(
        credentials.token("account-a").await,
        Err(CredentialError::Unavailable)
    );
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn denied_sign_out_invalidates_token_but_keeps_explicit_account_recovery_available() {
    let (credentials, backend) = setup();
    credentials.token("account-a").await.unwrap();
    *backend.delete_error.lock().unwrap() = Some(CredentialError::Unavailable);
    assert_eq!(
        credentials.remove("account-a").await,
        Err(CredentialError::Unavailable)
    );
    assert_eq!(
        credentials.token("account-a").await,
        Err(CredentialError::Unavailable)
    );
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
    credentials.retry_failed("account-a").unwrap();
    assert_eq!(
        credentials.token("account-a").await.unwrap(),
        "fixture-session-a"
    );
    assert_eq!(backend.reads.load(Ordering::SeqCst), 2);
    *backend.delete_error.lock().unwrap() = None;
    credentials.remove("account-a").await.unwrap();
    assert_eq!(
        credentials.token("account-a").await,
        Err(CredentialError::AccountChanged)
    );
}

#[tokio::test]
async fn missing_and_invalid_entries_are_distinct_from_access_denial_and_are_cached() {
    let (credentials, backend) = setup();
    backend.set("development", "account-a", Err(CredentialError::Missing));
    assert_eq!(
        credentials.token("account-a").await,
        Err(CredentialError::Missing)
    );
    assert_eq!(
        credentials.token("account-a").await,
        Err(CredentialError::Missing)
    );
    assert_eq!(backend.reads.load(Ordering::SeqCst), 1);
    backend.set("development", "account-a", Ok("invalid;session"));
    credentials.retry_failed("account-a").unwrap();
    assert_eq!(
        credentials.token("account-a").await,
        Err(CredentialError::Invalid)
    );
    assert_eq!(
        storage_error(keyring::Error::NoEntry),
        CredentialError::Missing
    );
    assert_eq!(
        storage_error(keyring::Error::PlatformFailure(Box::new(
            std::io::Error::from(std::io::ErrorKind::PermissionDenied)
        ))),
        CredentialError::Unavailable
    );
    assert_ne!(
        CredentialError::Missing.message(),
        CredentialError::Unavailable.message()
    );
}
