pub mod activation;
pub mod auth;
pub mod background;
pub mod collaboration;
pub mod commands;
mod credentials;
pub mod download;
pub mod drafts;
pub mod files;
pub mod identity;
pub mod media;
pub mod model;
pub mod native;
pub mod notifications;
pub mod recovery;
pub mod settings;
pub mod store;
pub mod sync;
#[cfg(test)]
mod tests;
pub mod transport;
pub mod window;

use std::sync::Arc;
use store::Store;
use tauri::{Manager, WebviewWindow};
use transport::Transport;

pub struct Runtime {
    pub store: Store,
    pub transport: Transport,
    pub sync: tokio::sync::Mutex<()>,
    pub login: std::sync::Mutex<Option<auth::Login>>,
    pub allow_exit: std::sync::atomic::AtomicBool,
}

/// Bind every account-scoped IPC request to the session that created it.
/// Delayed renderer work must never silently acquire the next signed-in user.
pub fn require_account(runtime: &Runtime, expected: &str) -> Result<String, String> {
    let db = runtime.store.0.lock().map_err(|error| error.to_string())?;
    require_account_in(&db, expected)?;
    Ok(expected.to_owned())
}

/// Also check under the write transaction to prevent a logout racing a save.
pub fn require_account_in(db: &rusqlite::Connection, expected: &str) -> Result<(), String> {
    use rusqlite::OptionalExtension;
    let encoded: Option<String> = db
        .query_row(
            "SELECT value FROM settings WHERE key='account'",
            [],
            |row| row.get(0),
        )
        .optional()
        .map_err(|error| error.to_string())?;
    let current = encoded.and_then(|value| serde_json::from_str::<String>(&value).ok());
    if expected.is_empty() || current.as_deref() != Some(expected) {
        return Err("Your account changed. Reopen this screen before continuing.".into());
    }
    Ok(())
}

pub fn runtime(window: &WebviewWindow) -> Result<Arc<Runtime>, String> {
    let url = window.url().map_err(|e| e.to_string())?;
    let local = is_local_url(&url);
    let development =
        cfg!(debug_assertions) && url.origin().ascii_serialization() == "http://localhost:5178";
    if window.label() != "main" || (!local && !development) {
        return Err("Desktop data is available only to the bundled application".into());
    }
    Ok(window.state::<Arc<Runtime>>().inner().clone())
}

pub fn is_local_url(url: &tauri::Url) -> bool {
    url.username().is_empty()
        && url.password().is_none()
        && url.port().is_none()
        && ((url.scheme() == "tauri" && url.host_str() == Some("localhost"))
            || (matches!(url.scheme(), "http" | "https")
                && url.host_str() == Some("tauri.localhost")))
}
