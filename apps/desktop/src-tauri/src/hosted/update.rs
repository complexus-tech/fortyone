//! Explicit signed updates for the full Projects shell. No endpoint or signing
//! key can be supplied by the renderer.
use serde::{Deserialize, Serialize};
use std::{
    sync::atomic::{AtomicBool, Ordering},
    time::{Duration, Instant},
};
use tauri::{Manager, WebviewWindow};
use tauri_plugin_updater::UpdaterExt;

static INSTALLING: AtomicBool = AtomicBool::new(false);

pub(super) fn in_progress() -> bool {
    INSTALLING.load(Ordering::Acquire)
}

struct Installation;
impl Installation {
    fn begin() -> Result<Self, String> {
        INSTALLING
            .compare_exchange(false, true, Ordering::AcqRel, Ordering::Acquire)
            .map(|_| Self)
            .map_err(|_| "An update is already in progress".into())
    }
}
impl Drop for Installation {
    fn drop(&mut self) {
        INSTALLING.store(false, Ordering::Release);
    }
}

fn configured(production: bool, key: &str, endpoint: &str) -> bool {
    production
        && !key.trim().is_empty()
        && tauri::Url::parse(endpoint).is_ok_and(|url| {
            url.scheme() == "https"
                && url.host_str().is_some()
                && url.username().is_empty()
                && url.password().is_none()
        })
}
fn enabled() -> bool {
    configured(
        cfg!(feature = "production"),
        option_env!("FORTYONE_UPDATER_PUBLIC_KEY").unwrap_or_default(),
        option_env!("FORTYONE_UPDATER_URL").unwrap_or_default(),
    )
}
fn updater(window: &WebviewWindow) -> Result<tauri_plugin_updater::Updater, String> {
    if !enabled() {
        return Err("Updates are not configured for this build".into());
    }
    window
        .updater_builder()
        .pubkey(option_env!("FORTYONE_UPDATER_PUBLIC_KEY").unwrap_or_default())
        .endpoints(vec![tauri::Url::parse(
            option_env!("FORTYONE_UPDATER_URL").unwrap_or_default(),
        )
        .map_err(|error| error.to_string())?])
        .map_err(|error| error.to_string())?
        .timeout(Duration::from_secs(120))
        .build()
        .map_err(|error| error.to_string())
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct UpdateStatus {
    configured: bool,
    current_version: &'static str,
}

#[derive(Serialize)]
pub struct AvailableUpdate {
    available: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    version: Option<String>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub struct InstallRequest {
    version: String,
}

#[tauri::command]
pub fn hosted_get_update_status(window: WebviewWindow) -> Result<UpdateStatus, String> {
    super::authorize_window(&window)?;
    Ok(UpdateStatus {
        configured: enabled(),
        current_version: env!("CARGO_PKG_VERSION"),
    })
}

#[tauri::command]
pub async fn hosted_check_update(window: WebviewWindow) -> Result<AvailableUpdate, String> {
    super::authorize_window(&window)?;
    let update = updater(&window)?
        .check()
        .await
        .map_err(|error| error.to_string())?;
    super::authorize_window(&window)?;
    Ok(AvailableUpdate {
        available: update.is_some(),
        version: update.map(|value| value.version),
    })
}

fn progress(window: &WebviewWindow, phase: &str, received: u64, total: Option<u64>) {
    let detail = serde_json::json!({"phase":phase,"received":received,"total":total});
    let _=window.eval(format!("window.dispatchEvent(new CustomEvent('fortyone:desktop-update-progress',{{detail:{detail}}}))"));
}

#[tauri::command]
pub async fn hosted_install_update(
    window: WebviewWindow,
    request: InstallRequest,
) -> Result<(), String> {
    super::authorize_window(&window)?;
    let _installation = Installation::begin()?;
    let result = install(&window, request).await;
    if let Err(message) = &result {
        super::close::failure(&window, message);
    }
    result
}

async fn install(window: &WebviewWindow, request: InstallRequest) -> Result<(), String> {
    let state = super::authorize_window(&window)?;
    if request.version.is_empty() || request.version.len() > 128 {
        return Err("Select the update version before installing".into());
    }
    let update = updater(&window)?
        .check()
        .await
        .map_err(|error| error.to_string())?
        .ok_or("No update is currently available")?;
    if update.version != request.version {
        return Err(
            "The available version changed. Check for updates again before installing.".into(),
        );
    }
    let mut received = 0u64;
    let mut last = Instant::now();
    progress(&window, "downloading", 0, None);
    // The updater verifies the package signature as part of download().
    let bytes = update
        .download(
            |chunk, total| {
                received = received.saturating_add(chunk as u64);
                if last.elapsed() >= Duration::from_millis(250) {
                    progress(&window, "downloading", received, total);
                    last = Instant::now();
                }
            },
            || {},
        )
        .await
        .map_err(|error| error.to_string())?;
    super::authorize_window(&window)?;
    progress(&window, "saving", received, Some(received));
    super::close::flush_for_update(&window).await?;
    state
        .db
        .lock()
        .map_err(|error| error.to_string())?
        .execute_batch("PRAGMA wal_checkpoint(FULL)")
        .map_err(|error| error.to_string())?;
    progress(&window, "installing", received, Some(received));
    super::close::allow_installer_exit(window.app_handle(), true)?;
    let result = tauri::async_runtime::spawn_blocking(move || update.install(bytes)).await;
    let installed = match result {
        Ok(result) => result.map_err(|error| error.to_string()),
        Err(error) => Err(error.to_string()),
    };
    if let Err(error) = installed {
        super::close::allow_installer_exit(window.app_handle(), false)?;
        return Err(error);
    }
    window.app_handle().restart()
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn base_configuration_initializes_the_plugin_without_enabling_updates() {
        let app: serde_json::Value =
            serde_json::from_str(include_str!("../../tauri.conf.json")).unwrap();
        let updater: tauri_plugin_updater::Config =
            serde_json::from_value(app["plugins"]["updater"].clone()).unwrap();
        assert!(updater.pubkey.is_empty());
        assert!(updater.endpoints.is_empty());
        assert!(!configured(false, &updater.pubkey, ""));
    }

    #[test]
    fn only_explicit_release_configuration_enables_updates() {
        assert!(configured(
            true,
            "signing-key",
            "https://updates.fortyone.app/latest.json"
        ));
        for (production, key, url) in [
            (false, "key", "https://updates.example/latest"),
            (true, "", "https://updates.example/latest"),
            (true, "key", "http://updates.example/latest"),
            (true, "key", "https://user:password@updates.example/latest"),
            (true, "key", "invalid"),
        ] {
            assert!(!configured(production, key, url));
        }
    }
    #[test]
    fn failed_or_completed_installation_releases_the_single_flight_guard() {
        let guard = Installation::begin().unwrap();
        assert!(in_progress());
        assert!(Installation::begin().is_err());
        drop(guard);
        assert!(!in_progress());
        assert!(Installation::begin().is_ok());
    }
}
