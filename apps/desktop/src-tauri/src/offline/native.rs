use super::{model::Settings, runtime};
use serde_json::{json, Value};
use tauri::{Emitter, Manager, WebviewWindow};
use tauri_plugin_autostart::ManagerExt;
use tauri_plugin_notification::NotificationExt;
use tauri_plugin_updater::UpdaterExt;

fn update_configured() -> bool {
    option_env!("FORTYONE_UPDATER_PUBLIC_KEY").is_some_and(|v| !v.is_empty())
        && option_env!("FORTYONE_UPDATER_URL").is_some_and(|v| v.starts_with("https://"))
}

#[tauri::command]
pub fn desktop_settings(window: WebviewWindow) -> Result<Settings, String> {
    let runtime = runtime(&window)?;
    let mut settings: Settings = runtime
        .store
        .setting("preferences")?
        .map(serde_json::from_value)
        .transpose()
        .map_err(|e| e.to_string())?
        .unwrap_or_default();
    settings.app_version = env!("CARGO_PKG_VERSION").into();
    settings.update_configured = update_configured();
    Ok(settings)
}

#[tauri::command]
pub fn desktop_update_settings(
    window: WebviewWindow,
    mut settings: Settings,
) -> Result<Settings, String> {
    let runtime = runtime(&window)?;
    let previous: Settings = runtime
        .store
        .setting("preferences")?
        .map(serde_json::from_value)
        .transpose()
        .map_err(|e| e.to_string())?
        .unwrap_or_default();
    if !matches!(settings.theme.as_str(), "light" | "dark" | "system") {
        return Err("Invalid theme".into());
    }
    if settings.notifications && !previous.notifications {
        let permission = window
            .notification()
            .request_permission()
            .map_err(|e| e.to_string())?;
        if permission != tauri_plugin_notification::PermissionState::Granted {
            return Err("Allow notifications in your system settings first".into());
        }
    }
    if settings.launch_at_login != previous.launch_at_login {
        let autostart = window.app_handle().autolaunch();
        if settings.launch_at_login {
            autostart.enable()
        } else {
            autostart.disable()
        }
        .map_err(|e| e.to_string())?;
    }
    window
        .set_theme(match settings.theme.as_str() {
            "light" => Some(tauri::Theme::Light),
            "dark" => Some(tauri::Theme::Dark),
            _ => None,
        })
        .map_err(|e| e.to_string())?;
    settings.update_configured = update_configured();
    settings.app_version = env!("CARGO_PKG_VERSION").into();
    runtime.store.set_setting(
        "preferences",
        &serde_json::to_value(&settings).map_err(|e| e.to_string())?,
    )?;
    window
        .emit("desktop-state-changed", ())
        .map_err(|e| e.to_string())?;
    Ok(settings)
}

#[tauri::command]
pub fn desktop_confirm_exit(window: WebviewWindow) -> Result<(), String> {
    let runtime = runtime(&window)?;
    runtime
        .store
        .0
        .lock()
        .map_err(|e| e.to_string())?
        .execute_batch("PRAGMA wal_checkpoint(FULL)")
        .map_err(|e| e.to_string())?;
    runtime
        .allow_exit
        .store(true, std::sync::atomic::Ordering::SeqCst);
    window.app_handle().exit(0);
    Ok(())
}

pub fn updater(window: &WebviewWindow) -> Result<tauri_plugin_updater::Updater, String> {
    if !update_configured() {
        return Err("Updates are not configured for this build".into());
    }
    window
        .updater_builder()
        .pubkey(option_env!("FORTYONE_UPDATER_PUBLIC_KEY").unwrap_or_default())
        .endpoints(vec![tauri::Url::parse(
            option_env!("FORTYONE_UPDATER_URL").unwrap_or_default(),
        )
        .map_err(|e| e.to_string())?])
        .map_err(|e| e.to_string())?
        .build()
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn desktop_check_update(window: WebviewWindow) -> Result<Value, String> {
    runtime(&window)?;
    let update = updater(&window)?.check().await.map_err(|e| e.to_string())?;
    Ok(match update {
        Some(update) => json!({"available":true,"version":update.version}),
        None => json!({"available":false}),
    })
}

#[tauri::command]
pub async fn desktop_install_update(window: WebviewWindow) -> Result<(), String> {
    use tauri::Emitter;
    let runtime = runtime(&window)?;
    let _guard = runtime.sync.lock().await;
    // SQLite WAL uses FULL synchronization. Checkpoint before a native installer
    // can replace the process; the outbox and drafts stay on disk for restart.
    runtime
        .store
        .0
        .lock()
        .map_err(|e| e.to_string())?
        .execute_batch("PRAGMA wal_checkpoint(FULL)")
        .map_err(|e| e.to_string())?;
    let Some(update) = updater(&window)?.check().await.map_err(|e| e.to_string())? else {
        return Ok(());
    };
    let app = window.app_handle().clone();
    update
        .download_and_install(
            move |received, total| {
                let _ = app.emit(
                    "desktop-update-progress",
                    json!({"received":received,"total":total}),
                );
            },
            || {},
        )
        .await
        .map_err(|e| e.to_string())?;
    window.app_handle().restart()
}
