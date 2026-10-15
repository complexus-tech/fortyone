use rusqlite::{params, OptionalExtension};
use serde::{Deserialize, Serialize};
use tauri::{
    menu::{Menu, MenuItem},
    tray::TrayIconBuilder,
    AppHandle, Manager, WebviewWindow,
};
use tauri_plugin_autostart::ManagerExt;

use super::{authorize_window, State};

use super::close;
pub use close::prevent_exit;

#[derive(Clone, Debug, Deserialize, Serialize, PartialEq)]
#[serde(rename_all = "camelCase", default)]
pub struct Preferences {
    pub notifications: bool,
    pub keep_running: bool,
    pub autostart: bool,
}

impl Default for Preferences {
    fn default() -> Self {
        Self {
            notifications: false,
            keep_running: true,
            autostart: false,
        }
    }
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct PreferencePatch {
    notifications: Option<bool>,
    keep_running: Option<bool>,
    autostart: Option<bool>,
}

pub fn read_preferences(state: &State) -> Result<Preferences, String> {
    let db = state.db.lock().map_err(|error| error.to_string())?;
    read(&db)
}

pub(super) fn read(db: &rusqlite::Connection) -> Result<Preferences, String> {
    let stored: Option<String> = db
        .query_row(
            "SELECT value FROM hosted_preferences WHERE key='preferences'",
            [],
            |row| row.get(0),
        )
        .optional()
        .map_err(|error| error.to_string())?;
    stored
        .map(|encoded| serde_json::from_str(&encoded).map_err(|error| error.to_string()))
        .transpose()
        .map(|value| value.unwrap_or_default())
}

#[tauri::command]
pub fn hosted_get_preferences(window: WebviewWindow) -> Result<Preferences, String> {
    let state = authorize_window(&window)?;
    let mut preferences = read_preferences(&state)?;
    preferences.autostart = window
        .autolaunch()
        .is_enabled()
        .map_err(|error| error.to_string())?;
    Ok(preferences)
}

#[tauri::command]
pub fn hosted_update_preferences(
    window: WebviewWindow,
    preferences: PreferencePatch,
) -> Result<Preferences, String> {
    let state = authorize_window(&window)?;
    // Serialize read/merge/write so independent settings controls cannot clobber one another.
    let db = state.db.lock().map_err(|error| error.to_string())?;
    let mut next = read(&db)?;
    if let Some(enabled) = preferences.notifications {
        next.notifications = enabled;
    }
    if let Some(enabled) = preferences.keep_running {
        next.keep_running = enabled;
    }
    let previous_autostart = window
        .autolaunch()
        .is_enabled()
        .map_err(|error| error.to_string())?;
    next.autostart = preferences.autostart.unwrap_or(previous_autostart);
    if next.autostart != previous_autostart {
        let autostart = window.autolaunch();
        if next.autostart {
            autostart.enable()
        } else {
            autostart.disable()
        }
        .map_err(|error| error.to_string())?;
    }
    let encoded = serde_json::to_string(&next).map_err(|error| error.to_string())?;
    if let Err(error) = db.execute("INSERT INTO hosted_preferences(key,value) VALUES('preferences',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value",params![encoded]) {
        if next.autostart != previous_autostart {
            let autostart = window.autolaunch();
            let _ = if previous_autostart { autostart.enable() } else { autostart.disable() };
        }
        return Err(error.to_string());
    }
    Ok(next)
}

#[tauri::command]
pub async fn hosted_request_notification_permission(
    window: WebviewWindow,
) -> Result<String, String> {
    authorize_window(&window)?;
    let permission = super::notifications::request_permission(&window).await?;
    authorize_window(&window)?;
    Ok(permission)
}

pub fn show(app: &AppHandle) {
    if let Some(window) = app.get_webview_window("main") {
        let _ = window.show();
        let _ = window.unminimize();
        let _ = window.set_focus();
    }
}

pub fn install(app: &AppHandle, window: &WebviewWindow) -> tauri::Result<()> {
    app.manage(close::CloseState::default());
    let app_handle = app.clone();
    let clone = window.clone();
    window.on_window_event(move |event| {
        if let tauri::WindowEvent::CloseRequested { api, .. } = event {
            api.prevent_close();
            let state = app_handle.state::<std::sync::Arc<State>>();
            let keep_running = read_preferences(state.inner())
                .map(|value| value.keep_running)
                .unwrap_or(true);
            close::request(
                &clone,
                if keep_running {
                    close::Action::Hide
                } else {
                    close::Action::Exit
                },
            );
        }
    });
    let menu = Menu::with_items(
        app,
        &[
            &MenuItem::with_id(app, "hosted-tray-open", "Open FortyOne", true, None::<&str>)?,
            &MenuItem::with_id(app, "hosted-tray-new", "New Story…", true, None::<&str>)?,
            &MenuItem::with_id(app, "hosted-tray-search", "Search…", true, None::<&str>)?,
            &MenuItem::with_id(app, "hosted-tray-quit", "Quit FortyOne", true, None::<&str>)?,
        ],
    )?;
    let builder = TrayIconBuilder::with_id("fortyone")
        .tooltip("FortyOne")
        .menu(&menu)
        .on_menu_event(|app, event| match event.id().as_ref() {
            "hosted-tray-open" => show(app),
            "hosted-tray-new" | "hosted-tray-search" => {
                show(app);
                if let Some(window) = app.get_webview_window("main") {
                    let action = if event.id().as_ref() == "hosted-tray-new" {
                        "create-story"
                    } else {
                        "search"
                    };
                    if let Err(error) = crate::chrome::dispatch(&window, action) {
                        eprintln!("Could not perform tray action: {error}");
                    }
                }
            }
            "hosted-tray-quit" => {
                prevent_exit(app);
            }
            _ => {}
        });
    crate::tray_icon::configure_icon(builder, app).build(app)?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn defaults_are_opt_in_for_notifications_and_login_launch() {
        let value = Preferences::default();
        assert!(!value.notifications);
        assert!(!value.autostart);
        assert!(value.keep_running);
    }
    #[test]
    fn reads_saved_preferences_and_rejects_unknown_patch_fields() {
        let db = rusqlite::Connection::open_in_memory().unwrap();
        db.execute_batch(
            "CREATE TABLE hosted_preferences(key TEXT PRIMARY KEY,value TEXT NOT NULL);",
        )
        .unwrap();
        assert_eq!(read(&db).unwrap(), Preferences::default());
        db.execute(
            "INSERT INTO hosted_preferences VALUES('preferences',?)",
            [r#"{"notifications":true,"keepRunning":false,"autostart":true}"#],
        )
        .unwrap();
        assert_eq!(
            read(&db).unwrap(),
            Preferences {
                notifications: true,
                keep_running: false,
                autostart: true
            }
        );
        assert!(serde_json::from_str::<PreferencePatch>(r#"{"unknown":true}"#).is_err());
    }
}
