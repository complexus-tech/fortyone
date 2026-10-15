use serde_json::Value;
use std::sync::Arc;
use tauri::{AppHandle, Emitter, Manager, WebviewWindow};

use super::{runtime, Runtime};

fn validate(runtime: &Runtime, target: &Value) -> Result<(), String> {
    let account = runtime.store.account()?;
    if target
        .get("accountId")
        .and_then(Value::as_str)
        .is_some_and(|id| id != account)
    {
        return Err("This link belongs to another account".into());
    }
    let workspace = target["workspaceId"]
        .as_str()
        .ok_or("Invalid workspace link")?;
    runtime.store.authorize(&account, workspace)?;
    match target["kind"].as_str() {
        Some("story" | "document") => {
            uuid::Uuid::parse_str(target["resourceId"].as_str().ok_or("Missing linked item")?)
                .map_err(|_| "Invalid linked item")?;
        }
        Some("notifications") => {}
        _ => return Err("Unsupported desktop destination".into()),
    }
    Ok(())
}

pub fn activate(app: &AppHandle, mut target: Value) -> Result<(), String> {
    let runtime = app.state::<Arc<Runtime>>();
    validate(&runtime, &target)?;
    target["accountId"] = Value::String(runtime.store.account()?);
    runtime
        .store
        .set_setting("workspace", &target["workspaceId"])?;
    runtime.store.set_setting("pendingNavigation", &target)?;
    app.emit("desktop-navigation", target)
        .map_err(|e| e.to_string())?;
    if let Some(window) = app.get_webview_window("main") {
        window.show().map_err(|e| e.to_string())?;
        window.set_focus().map_err(|e| e.to_string())?;
    }
    Ok(())
}

#[tauri::command]
pub fn desktop_take_navigation(
    window: WebviewWindow,
    expected_account_id: String,
) -> Result<Option<Value>, String> {
    use rusqlite::OptionalExtension;
    let runtime = runtime(&window)?;
    super::require_account(&runtime, &expected_account_id)?;
    let target = {
        let mut db = runtime.store.0.lock().map_err(|error| error.to_string())?;
        let tx = db.transaction().map_err(|error| error.to_string())?;
        super::require_account_in(&tx, &expected_account_id)?;
        let encoded: Option<String> = tx
            .query_row(
                "SELECT value FROM settings WHERE key='pendingNavigation'",
                [],
                |row| row.get(0),
            )
            .optional()
            .map_err(|error| error.to_string())?;
        tx.execute("DELETE FROM settings WHERE key='pendingNavigation'", [])
            .map_err(|error| error.to_string())?;
        tx.commit().map_err(|error| error.to_string())?;
        encoded
            .map(|value| serde_json::from_str::<Value>(&value).map_err(|error| error.to_string()))
            .transpose()?
            .filter(|value| !value.is_null())
    };
    if let Some(target) = &target {
        validate(&runtime, target)?;
    }
    Ok(target)
}
