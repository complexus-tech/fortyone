//! Narrow native capabilities for the full hosted Projects application.
//! Browser session cookies never cross IPC; persisted bindings contain hashes.
pub mod attachments;
pub mod cache;
pub mod close;
pub mod drafts;
pub mod export;
mod identity;
pub mod notifications;
pub mod outbox;
pub mod pages;
pub mod preferences;
#[cfg(test)]
mod tests;
pub mod update;

use rusqlite::{params, Connection, OptionalExtension};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use sha2::{Digest, Sha256};
use std::{
    path::Path,
    sync::{
        atomic::{AtomicU64, Ordering},
        Arc, Mutex,
    },
    time::{Duration, SystemTime, UNIX_EPOCH},
};
use tauri::{Manager, Url, WebviewWindow};

const LEASE_MS: u64 = 7 * 24 * 3600 * 1000;

pub struct State {
    pub(super) db: Mutex<Connection>,
    pub(super) client: reqwest::Client,
    binding: tokio::sync::Mutex<()>,
    generation: AtomicU64,
    outbox: tokio::sync::Mutex<()>,
}

pub struct BoundScope {
    pub origin: String,
    pub account: String,
    pub workspace: String,
    pub slug: String,
    pub cookie_hash: String,
    // Native HTTP only. This type must never implement Serialize or Debug.
    pub token: String,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct SessionRequest {
    account_id: String,
    workspace_id: String,
    workspace_slug: String,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct RestoreRequest {
    workspace_slug: String,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct RestoredSession {
    account_id: String,
    workspace_id: String,
    workspace_slug: String,
    expires_at: u64,
}

pub(super) fn now() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_millis() as u64
}

pub fn open(path: &Path) -> Result<State, String> {
    let db = Connection::open(path).map_err(|error| error.to_string())?;
    initialize(&db)?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        std::fs::set_permissions(path, std::fs::Permissions::from_mode(0o600))
            .map_err(|error| error.to_string())?;
    }
    Ok(State {
        db: Mutex::new(db),
        client: reqwest::Client::builder()
            .redirect(reqwest::redirect::Policy::none())
            .timeout(Duration::from_secs(15))
            .connect_timeout(Duration::from_secs(5))
            .build()
            .map_err(|error| error.to_string())?,
        binding: tokio::sync::Mutex::new(()),
        generation: AtomicU64::new(0),
        outbox: tokio::sync::Mutex::new(()),
    })
}

fn initialize(db: &Connection) -> Result<(), String> {
    db.busy_timeout(Duration::from_secs(5))
        .map_err(|error| error.to_string())?;
    db.execute_batch("PRAGMA journal_mode=WAL; PRAGMA synchronous=FULL; PRAGMA secure_delete=ON; PRAGMA max_page_count=65536;
      CREATE TABLE IF NOT EXISTS hosted_bindings(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,slug TEXT NOT NULL,cookie_hash TEXT NOT NULL,verified INTEGER NOT NULL,PRIMARY KEY(origin,account,workspace));
      CREATE TABLE IF NOT EXISTS hosted_active(origin TEXT PRIMARY KEY,account TEXT NOT NULL,cookie_hash TEXT NOT NULL);
      CREATE TABLE IF NOT EXISTS hosted_cache(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,version INTEGER NOT NULL,payload TEXT NOT NULL,saved INTEGER NOT NULL,PRIMARY KEY(origin,account,workspace));
      CREATE TABLE IF NOT EXISTS hosted_known(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,slug TEXT NOT NULL,cookie_hash TEXT NOT NULL,PRIMARY KEY(origin,account,workspace));
      CREATE TABLE IF NOT EXISTS hosted_drafts(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,key TEXT NOT NULL,payload TEXT NOT NULL,saved INTEGER NOT NULL,PRIMARY KEY(origin,account,workspace,key));
      CREATE TABLE IF NOT EXISTS hosted_operations(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,id TEXT NOT NULL,operation TEXT NOT NULL,PRIMARY KEY(origin,account,workspace,id));
      CREATE TABLE IF NOT EXISTS hosted_aliases(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,provisional TEXT NOT NULL,resource TEXT NOT NULL,kind TEXT NOT NULL,version TEXT,creation_hash TEXT NOT NULL,PRIMARY KEY(origin,account,workspace,provisional));
      CREATE TABLE IF NOT EXISTS hosted_results(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,id TEXT NOT NULL,result TEXT NOT NULL,saved INTEGER NOT NULL,PRIMARY KEY(origin,account,workspace,id));
      CREATE TABLE IF NOT EXISTS hosted_preferences(key TEXT PRIMARY KEY,value TEXT NOT NULL);
      CREATE TABLE IF NOT EXISTS hosted_notifications(account TEXT NOT NULL,workspace TEXT NOT NULL,id TEXT NOT NULL,version TEXT NOT NULL,PRIMARY KEY(account,workspace,id,version));").map_err(|error| error.to_string())?;
    pages::initialize(db)?;
    attachments::initialize(db)?;
    identity::initialize(db)?;
    outbox::initialize(db)
}

pub fn authorize_window(window: &WebviewWindow) -> Result<Arc<State>, String> {
    let policy = window.state::<Arc<crate::navigation::NavigationPolicy>>();
    crate::chrome::authorize_window(window, &policy)?;
    Ok(window.state::<Arc<State>>().inner().clone())
}

fn valid_slug(slug: &str) -> bool {
    !slug.is_empty()
        && slug.len() <= 63
        && !slug.starts_with('-')
        && !slug.ends_with('-')
        && slug
            .bytes()
            .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == b'-')
}

fn origin_scope(policy: &crate::navigation::NavigationPolicy, url: &Url, slug: &str) -> bool {
    if !valid_slug(slug) || !policy.allows_navigation(url) {
        return false;
    }
    if policy.app_url().host_str() == Some("cloud.fortyone.app") {
        url.host_str() == Some(format!("{slug}.fortyone.app").as_str())
    } else {
        url.origin() == policy.app_url().origin() && url.path().split('/').nth(1) == Some(slug)
    }
}

fn caller_origin(window: &WebviewWindow, slug: &str) -> Result<String, String> {
    authorize_window(window)?;
    let url = window.url().map_err(|error| error.to_string())?;
    if !origin_scope(
        &window.state::<Arc<crate::navigation::NavigationPolicy>>(),
        &url,
        slug,
    ) {
        return Err("This native workspace does not match the current page".into());
    }
    Ok(url.origin().ascii_serialization())
}

pub(super) fn api_origin() -> &'static str {
    if cfg!(feature = "production") {
        "https://api.fortyone.app"
    } else {
        "http://localhost:8000"
    }
}

async fn browser_token(window: &WebviewWindow) -> Result<String, String> {
    authorize_window(window)?;
    let window = window.clone();
    // A fixed API URL prevents hosted content from requesting arbitrary cookies.
    // Wry cookie access may block and must not execute on the WebView UI thread.
    tauri::async_runtime::spawn_blocking(move || {
        // Wry 0.57 on macOS matches cookie domains by equality, incorrectly
        // excluding our .fortyone.app cookie from api.fortyone.app. Enumerate
        // natively there, then apply the fixed API's narrow policy below.
        #[cfg(target_os = "macos")]
        let cookies = window.cookies();
        #[cfg(not(target_os = "macos"))]
        let cookies = window
            .cookies_for_url(Url::parse(api_origin()).map_err(|_| "Invalid native API address")?);
        let cookies = cookies.map_err(|_| "Could not access the current browser session")?;
        select_session_cookie(cookies, api_origin(), now())
    })
    .await
    .map_err(|_| "Could not access the current browser session".to_owned())?
}

fn select_session_cookie(
    cookies: Vec<tauri::webview::Cookie<'static>>,
    api: &str,
    time: u64,
) -> Result<String, String> {
    let production = match api {
        "https://api.fortyone.app" => true,
        "http://localhost:8000" => false,
        _ => return Err("Invalid native API address".into()),
    };
    let mut sessions = cookies.into_iter().filter(|cookie| {
        // Cookie::domain() normalizes a leading dot. Accept only the API host
        // and our one trusted parent domain, never arbitrary suffix matches.
        let allowed_domain = cookie.domain().is_some_and(|domain| {
            if production {
                domain.eq_ignore_ascii_case("api.fortyone.app")
                    || domain.eq_ignore_ascii_case("fortyone.app")
            } else {
                domain.eq_ignore_ascii_case("localhost")
            }
        });
        cookie.name() == "fortyone_session"
            && cookie.http_only() == Some(true)
            && allowed_domain
            // The session covers every API endpoint; a path-limited cookie
            // must not be promoted to a root-scoped native credential.
            && cookie.path() == Some("/")
            && (!production || cookie.secure() == Some(true))
            && !cookie.max_age().is_some_and(|age| age.whole_seconds() <= 0)
            && !cookie.expires_datetime().is_some_and(|expires| {
                i128::from(expires.unix_timestamp()) * 1000 <= i128::from(time)
            })
    });
    let token = sessions
        .next()
        .map(|cookie| cookie.value().to_owned())
        .ok_or("Sign in to restore your downloaded workspace")?;
    if sessions.next().is_some() || token.is_empty() || token.contains(['\r', '\n', ';']) {
        return Err("The current browser session is invalid".into());
    }
    Ok(token)
}

fn fingerprint(token: &str) -> String {
    format!("{:x}", Sha256::digest(token.as_bytes()))
}

fn valid_lease(verified: u64, time: u64) -> bool {
    verified <= time.saturating_add(60_000) && time.saturating_sub(verified) <= LEASE_MS
}

fn binding(
    db: &Connection,
    origin: &str,
    account: &str,
    workspace: &str,
    hash: &str,
    time: u64,
) -> Result<(String, u64), String> {
    let record: Option<(String,u64)> = db.query_row(
        "SELECT slug,verified FROM hosted_bindings WHERE origin=?1 AND account=?2 AND workspace=?3 AND cookie_hash=?4 AND EXISTS(SELECT 1 FROM hosted_active WHERE origin=?1 AND account=?2 AND cookie_hash=?4)",
        params![origin,account,workspace,hash], |row| Ok((row.get(0)?,row.get(1)?)),
    ).optional().map_err(|error| error.to_string())?;
    let (slug, verified) =
        record.ok_or("Verify this workspace online before using its native cache")?;
    if !valid_lease(verified, time) {
        return Err("Reconnect to renew this workspace's offline access".into());
    }
    Ok((slug, verified))
}

pub(super) fn ensure_scope_current(db: &Connection, scope: &BoundScope) -> Result<(), String> {
    let (slug, _) = binding(
        db,
        &scope.origin,
        &scope.account,
        &scope.workspace,
        &scope.cookie_hash,
        now(),
    )?;
    if slug != scope.slug {
        return Err("Workspace identity changed".into());
    }
    Ok(())
}

pub async fn authorize_scope(
    window: &WebviewWindow,
    account: &str,
    workspace: &str,
    slug: Option<&str>,
) -> Result<BoundScope, String> {
    let state = authorize_window(window)?;
    let origin = window
        .url()
        .map_err(|error| error.to_string())?
        .origin()
        .ascii_serialization();
    let token = browser_token(window).await?;
    let hash = fingerprint(&token);
    let (bound_slug, _) = {
        let db = state.db.lock().map_err(|error| error.to_string())?;
        binding(&db, &origin, account, workspace, &hash, now())?
    };
    if slug.is_some_and(|value| value != bound_slug)
        || caller_origin(window, &bound_slug)? != origin
    {
        return Err("This native workspace does not match the current page".into());
    }
    Ok(BoundScope {
        origin,
        account: account.into(),
        workspace: workspace.into(),
        slug: bound_slug,
        cookie_hash: hash,
        token,
    })
}

enum VerifyError {
    Unavailable,
    Denied,
}

async fn get(state: &State, path: &str, token: &str) -> Result<Value, VerifyError> {
    let mut response = state
        .client
        .get(format!("{}{path}", api_origin()))
        .header(reqwest::header::COOKIE, format!("fortyone_session={token}"))
        .send()
        .await
        .map_err(|_| VerifyError::Unavailable)?;
    if matches!(response.status().as_u16(), 401 | 403 | 404) {
        return Err(VerifyError::Denied);
    }
    if !response.status().is_success() {
        return Err(VerifyError::Unavailable);
    }
    if response
        .content_length()
        .is_some_and(|size| size > 8 * 1024 * 1024)
    {
        return Err(VerifyError::Denied);
    }
    let mut body = Vec::new();
    while let Some(chunk) = response
        .chunk()
        .await
        .map_err(|_| VerifyError::Unavailable)?
    {
        if body.len().saturating_add(chunk.len()) > 8 * 1024 * 1024 {
            return Err(VerifyError::Denied);
        }
        body.extend_from_slice(&chunk);
    }
    let value: Value = serde_json::from_slice(&body).map_err(|_| VerifyError::Unavailable)?;
    Ok(value.get("data").cloned().unwrap_or(value))
}

fn forget_origin(db: &Connection, origin: &str) -> Result<(), String> {
    for table in [
        "hosted_bindings",
        "hosted_cache",
        "hosted_active",
        "hosted_results",
        "hosted_pages",
    ] {
        db.execute(&format!("DELETE FROM {table} WHERE origin=?"), [origin])
            .map_err(|error| error.to_string())?;
    }
    db.execute(
        "DELETE FROM hosted_files WHERE origin=? AND uploaded=1",
        [origin],
    )
    .map_err(|error| error.to_string())?;
    db.execute("DELETE FROM hosted_drafts WHERE origin=? AND key LIKE 'collaboration:%' AND json_extract(payload,'$.dirty')=0", [origin]).map_err(|error| error.to_string())?;
    Ok(())
}

#[tauri::command]
pub async fn hosted_bind_session(
    window: WebviewWindow,
    request: SessionRequest,
) -> Result<(), String> {
    let state = authorize_window(&window)?;
    let _guard = state.binding.lock().await;
    let generation = state.generation.load(Ordering::Acquire);
    uuid::Uuid::parse_str(&request.account_id).map_err(|_| "Invalid account identity")?;
    uuid::Uuid::parse_str(&request.workspace_id).map_err(|_| "Invalid workspace identity")?;
    let origin = caller_origin(&window, &request.workspace_slug)?;
    let token = browser_token(&window).await?;
    let hash = fingerprint(&token);
    let verified = async {
        let user = get(&state, "/auth/me", &token).await?;
        if user["id"].as_str() != Some(&request.account_id) {
            return Err(VerifyError::Denied);
        }
        let workspaces = get(&state, "/workspaces", &token).await?;
        if !workspaces.as_array().is_some_and(|values| {
            values.iter().any(|value| {
                value["id"].as_str() == Some(&request.workspace_id)
                    && value["slug"].as_str() == Some(&request.workspace_slug)
            })
        }) {
            return Err(VerifyError::Denied);
        }
        let workspace = get(
            &state,
            &format!("/workspaces/{}", request.workspace_slug),
            &token,
        )
        .await?;
        if workspace["id"].as_str() != Some(&request.workspace_id) {
            return Err(VerifyError::Denied);
        }
        Ok(())
    }
    .await;
    if fingerprint(&browser_token(&window).await?) != hash
        || caller_origin(&window, &request.workspace_slug)? != origin
    {
        return Err("Your account or workspace changed while verifying native access".into());
    }
    let mut db = state.db.lock().map_err(|error| error.to_string())?;
    if state.generation.load(Ordering::Acquire) != generation {
        return Err("Your session changed while verifying native access".into());
    }
    match verified {
        Err(VerifyError::Unavailable) => {
            let (slug, _) = binding(
                &db,
                &origin,
                &request.account_id,
                &request.workspace_id,
                &hash,
                now(),
            )?;
            if slug != request.workspace_slug {
                return Err("Workspace identity changed".into());
            }
            Ok(())
        }
        Err(VerifyError::Denied) => {
            forget_origin(&db, &origin)?;
            Err("Sign in and verify access before restoring this workspace".into())
        }
        Ok(()) => {
            let tx = db.transaction().map_err(|error| error.to_string())?;
            let previous: Option<(String, String)> = tx
                .query_row(
                    "SELECT account,cookie_hash FROM hosted_active WHERE origin=?",
                    [&origin],
                    |row| Ok((row.get(0)?, row.get(1)?)),
                )
                .optional()
                .map_err(|error| error.to_string())?;
            if previous
                .is_some_and(|(account, cookie)| account != request.account_id || cookie != hash)
            {
                forget_origin(&tx, &origin)?;
            }
            tx.execute("INSERT INTO hosted_active VALUES(?,?,?) ON CONFLICT(origin) DO UPDATE SET account=excluded.account,cookie_hash=excluded.cookie_hash",params![origin,request.account_id,hash]).map_err(|error|error.to_string())?;
            tx.execute("INSERT INTO hosted_bindings VALUES(?,?,?,?,?,?) ON CONFLICT(origin,account,workspace) DO UPDATE SET slug=excluded.slug,cookie_hash=excluded.cookie_hash,verified=excluded.verified",params![origin,request.account_id,request.workspace_id,request.workspace_slug,hash,now()]).map_err(|error|error.to_string())?;
            tx.execute("INSERT INTO hosted_known VALUES(?,?,?,?,?) ON CONFLICT(origin,account,workspace) DO UPDATE SET slug=excluded.slug,cookie_hash=excluded.cookie_hash",params![origin,request.account_id,request.workspace_id,request.workspace_slug,hash]).map_err(|error|error.to_string())?;
            tx.execute(
                "DELETE FROM hosted_cache WHERE saved<?",
                [now().saturating_sub(LEASE_MS)],
            )
            .map_err(|error| error.to_string())?;
            tx.commit().map_err(|error| error.to_string())
        }
    }
}

#[tauri::command]
pub async fn hosted_restore_session(
    window: WebviewWindow,
    request: RestoreRequest,
) -> Result<RestoredSession, String> {
    let state = authorize_window(&window)?;
    let origin = caller_origin(&window, &request.workspace_slug)?;
    let hash = fingerprint(&browser_token(&window).await?);
    if caller_origin(&window, &request.workspace_slug)? != origin {
        return Err("Workspace changed".into());
    }
    let db = state.db.lock().map_err(|error| error.to_string())?;
    let (account,workspace): (String,String)=db.query_row("SELECT account,workspace FROM hosted_bindings WHERE origin=? AND slug=? AND cookie_hash=? ORDER BY verified DESC LIMIT 1",params![origin,request.workspace_slug,hash],|row|Ok((row.get(0)?,row.get(1)?))).map_err(|_|"Verify this workspace online before restoring it")?;
    let (slug, verified) = binding(&db, &origin, &account, &workspace, &hash, now())?;
    Ok(RestoredSession {
        account_id: account,
        workspace_id: workspace,
        workspace_slug: slug,
        expires_at: verified.saturating_add(LEASE_MS),
    })
}
