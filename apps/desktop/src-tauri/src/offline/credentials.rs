//! OS credentials stay in native memory after one serialized secure-store read.
//! Failed access is cached as well: background work must never reopen a prompt.
use std::sync::{Arc, Mutex};

#[cfg(test)]
mod tests;

#[derive(Clone, Copy, Debug, PartialEq)]
pub(super) enum CredentialError {
    Missing,
    Unavailable,
    Invalid,
    AccountChanged,
}

impl CredentialError {
    pub fn message(self) -> String {
        match self {
            Self::Missing => "Sign in again to synchronize. Your local work is retained.",
            Self::Unavailable => "Secure credential access was not granted or is unavailable. Allow access to your system credential store, then choose Sync now to retry. Your local work is retained.",
            Self::Invalid => "Your saved desktop session is invalid. Sign in again. Your local work is retained.",
            Self::AccountChanged => "Your account changed. Reopen this screen before continuing.",
        }.into()
    }
}

fn storage_error(error: keyring::Error) -> CredentialError {
    match error {
        keyring::Error::NoEntry => CredentialError::Missing,
        keyring::Error::BadEncoding(_) => CredentialError::Invalid,
        // macOS reports canceled/denied dialogs as PlatformFailure, while other
        // providers use NoStorageAccess. Neither means the account is signed out.
        _ => CredentialError::Unavailable,
    }
}

trait CredentialBackend: Send + Sync {
    fn read(&self, environment: &str, account: &str) -> Result<String, CredentialError>;
    fn write(&self, environment: &str, account: &str, token: &str) -> Result<(), CredentialError>;
    fn delete(&self, environment: &str, account: &str) -> Result<(), CredentialError>;
}

struct SystemCredentials;

fn entry(environment: &str, account: &str) -> Result<keyring::Entry, CredentialError> {
    keyring::Entry::new("app.fortyone.desktop", &format!("{environment}:{account}"))
        .map_err(storage_error)
}

impl CredentialBackend for SystemCredentials {
    fn read(&self, environment: &str, account: &str) -> Result<String, CredentialError> {
        entry(environment, account)?
            .get_password()
            .map_err(storage_error)
    }
    fn write(&self, environment: &str, account: &str, token: &str) -> Result<(), CredentialError> {
        entry(environment, account)?
            .set_password(token)
            .map_err(storage_error)
    }
    fn delete(&self, environment: &str, account: &str) -> Result<(), CredentialError> {
        match entry(environment, account)?.delete_credential() {
            Ok(()) | Err(keyring::Error::NoEntry) => Ok(()),
            Err(error) => Err(storage_error(error)),
        }
    }
}

struct State {
    account: Option<String>,
    generation: u64,
    cached: Option<Result<String, CredentialError>>,
}

pub struct Credentials {
    environment: &'static str,
    backend: Arc<dyn CredentialBackend>,
    state: Arc<Mutex<State>>,
    io: Arc<tokio::sync::Mutex<()>>,
}

impl Credentials {
    pub fn new(environment: &'static str, account: Option<String>) -> Self {
        Self::with_backend(environment, account, Arc::new(SystemCredentials))
    }

    fn with_backend(
        environment: &'static str,
        account: Option<String>,
        backend: Arc<dyn CredentialBackend>,
    ) -> Self {
        Self {
            environment,
            backend,
            state: Arc::new(Mutex::new(State {
                account,
                generation: 0,
                cached: None,
            })),
            io: Arc::new(tokio::sync::Mutex::new(())),
        }
    }

    fn cached(
        &self,
        account: &str,
    ) -> Result<Option<Result<String, CredentialError>>, CredentialError> {
        let state = self
            .state
            .lock()
            .map_err(|_| CredentialError::Unavailable)?;
        if state.account.as_deref() != Some(account) || account.is_empty() {
            return Err(CredentialError::AccountChanged);
        }
        Ok(state.cached.clone())
    }

    pub async fn token(&self, account: &str) -> Result<String, CredentialError> {
        if let Some(cached) = self.cached(account)? {
            return cached;
        }
        let guard = self.io.clone().lock_owned().await;
        if let Some(cached) = self.cached(account)? {
            return cached;
        }
        let generation = self
            .state
            .lock()
            .map_err(|_| CredentialError::Unavailable)?
            .generation;
        let (backend, state) = (self.backend.clone(), self.state.clone());
        let (environment, account) = (self.environment, account.to_owned());
        // Move the guard into the blocking job: canceling an async caller cannot
        // permit a second prompt while the first OS dialog is still open.
        tokio::task::spawn_blocking(move || {
            let _guard = guard;
            let result = backend.read(environment, &account).and_then(validate_token);
            cache_result(&state, &account, generation, result)
        })
        .await
        .map_err(|_| CredentialError::Unavailable)?
    }

    /// Only an explicit user action may retry a previously denied OS prompt.
    pub fn retry_failed(&self, account: &str) -> Result<(), CredentialError> {
        let mut state = self
            .state
            .lock()
            .map_err(|_| CredentialError::Unavailable)?;
        if state.account.as_deref() != Some(account) {
            return Err(CredentialError::AccountChanged);
        }
        if state.cached.as_ref().is_some_and(Result::is_err) {
            state.generation = state.generation.wrapping_add(1);
            state.cached = None;
        }
        Ok(())
    }

    pub async fn replace(&self, account: &str, token: String) -> Result<(), CredentialError> {
        let token = validate_token(token)?;
        let guard = self.io.clone().lock_owned().await;
        let generation = {
            let mut state = self
                .state
                .lock()
                .map_err(|_| CredentialError::Unavailable)?;
            state.generation = state.generation.wrapping_add(1);
            state.account = Some(account.to_owned());
            state.cached = None;
            state.generation
        };
        let (backend, state) = (self.backend.clone(), self.state.clone());
        let (environment, account) = (self.environment, account.to_owned());
        tokio::task::spawn_blocking(move || {
            let _guard = guard;
            let result = backend.write(environment, &account, &token).map(|()| token);
            cache_result(&state, &account, generation, result).map(|_| ())
        })
        .await
        .map_err(|_| CredentialError::Unavailable)?
    }

    pub async fn remove(&self, account: &str) -> Result<(), CredentialError> {
        let generation = {
            let mut state = self
                .state
                .lock()
                .map_err(|_| CredentialError::Unavailable)?;
            if state
                .account
                .as_deref()
                .is_some_and(|current| current != account)
            {
                return Err(CredentialError::AccountChanged);
            }
            state.account = None;
            state.generation = state.generation.wrapping_add(1);
            state.cached = None;
            state.generation
        };
        let guard = self.io.clone().lock_owned().await;
        let (backend, state) = (self.backend.clone(), self.state.clone());
        let (environment, account) = (self.environment, account.to_owned());
        tokio::task::spawn_blocking(move || {
            let _guard = guard;
            let result = backend.delete(environment, &account);
            if let Err(error) = result {
                // Local sign-out cannot complete until the OS entry is removed.
                // Keep the account usable for explicit recovery without reviving
                // its cached token or a read that started before sign-out.
                let mut state = state.lock().map_err(|_| CredentialError::Unavailable)?;
                if state.generation == generation && state.account.is_none() {
                    state.account = Some(account);
                    state.cached = Some(Err(error));
                }
            }
            result
        })
        .await
        .map_err(|_| CredentialError::Unavailable)?
    }
}

fn validate_token(token: String) -> Result<String, CredentialError> {
    if token.is_empty() || token.contains(['\r', '\n', ';']) {
        Err(CredentialError::Invalid)
    } else {
        Ok(token)
    }
}

fn cache_result(
    state: &Mutex<State>,
    account: &str,
    generation: u64,
    result: Result<String, CredentialError>,
) -> Result<String, CredentialError> {
    let mut state = state.lock().map_err(|_| CredentialError::Unavailable)?;
    if state.account.as_deref() != Some(account) || state.generation != generation {
        return Err(CredentialError::AccountChanged);
    }
    state.cached = Some(result.clone());
    result
}
