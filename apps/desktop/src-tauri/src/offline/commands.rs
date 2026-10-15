use super::{model::Operation, runtime};
use rusqlite::params;
use serde_json::{json, Value};
use tauri::{Emitter, Manager, WebviewWindow};
use tauri_plugin_opener::OpenerExt;

#[tauri::command]
pub fn desktop_bootstrap(window: WebviewWindow) -> Result<Value, String> {
    runtime(&window)?.store.bootstrap()
}

#[tauri::command]
pub fn desktop_read_workspace(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
) -> Result<Value, String> {
    let runtime = runtime(&window)?;
    let account = super::require_account(&runtime, &expected_account_id)?;
    runtime.store.workspace(&account, &workspace_id)
}

#[tauri::command]
pub fn desktop_select_workspace(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
) -> Result<Value, String> {
    let runtime = runtime(&window)?;
    let account = super::require_account(&runtime, &expected_account_id)?;
    let snapshot = runtime.store.workspace(&account, &workspace_id)?;
    {
        let db = runtime.store.0.lock().map_err(|error| error.to_string())?;
        super::require_account_in(&db, &account)?;
        db.execute("INSERT INTO settings VALUES('workspace',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", [json!(workspace_id).to_string()]).map_err(|error|error.to_string())?;
    }
    window
        .emit("desktop-state-changed", ())
        .map_err(|e| e.to_string())?;
    Ok(snapshot)
}

#[tauri::command]
pub fn desktop_save_operation(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    kind: String,
    resource_id: String,
    payload: Value,
    base_version: Option<String>,
    consumed_draft: Option<super::identity::ConsumedDraft>,
) -> Result<Value, String> {
    let runtime = runtime(&window)?;
    let account = super::require_account(&runtime, &expected_account_id)?;
    if kind == "attachment.upload" {
        return Err("Choose an attachment through the file picker".into());
    }
    let operation = Operation {
        id: uuid::Uuid::new_v4().to_string(),
        kind,
        resource_id,
        payload,
        base_version,
        status: "pending".into(),
        error: None,
        server_version: None,
        server_value: None,
        attempted: false,
    };
    runtime.store.enqueue_consuming(
        &account,
        &workspace_id,
        &operation,
        consumed_draft.as_ref(),
    )?;
    runtime.store.workspace(&account, &workspace_id)
}

#[tauri::command]
pub async fn desktop_resolve_operation(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    operation_id: String,
    resolution: String,
    payload: Option<Value>,
    base_version: Option<String>,
) -> Result<Value, String> {
    let runtime = runtime(&window)?;
    let _guard = runtime.sync.lock().await;
    let account = super::require_account(&runtime, &expected_account_id)?;
    runtime.store.authorize(&account, &workspace_id)?;
    let mut operation = runtime
        .store
        .operations(&account, &workspace_id)?
        .into_iter()
        .find(|v| v.id == operation_id)
        .ok_or("This pending change no longer exists")?;
    match resolution.as_str() {
        "discard" => {
            if operation.status == "pending" && operation.attempted {
                return Err("This change may already be saved on the server. Reconnect and reconcile it before discarding.".into());
            }
            if operation.kind.ends_with(".create")
                && runtime
                    .store
                    .operations(&account, &workspace_id)?
                    .iter()
                    .any(|other| {
                        other.id != operation.id
                            && (other.resource_id == operation.resource_id
                                || ["storyId", "documentId", "parentId"].iter().any(|key| {
                                    other.payload[*key].as_str() == Some(&operation.resource_id)
                                }))
                    })
            {
                return Err(
                    "Resolve changes attached to this item before discarding its creation".into(),
                );
            }
            let mut db = runtime.store.0.lock().map_err(|e| e.to_string())?;
            let tx = db.transaction().map_err(|e| e.to_string())?;
            tx.execute(
                "DELETE FROM operations WHERE account=? AND workspace=? AND id=?",
                params![account, workspace_id, operation_id],
            )
            .map_err(|e| e.to_string())?;
            tx.execute(
                "DELETE FROM attachments WHERE account=? AND workspace=? AND id=?",
                params![account, workspace_id, operation_id],
            )
            .map_err(|e| e.to_string())?;
            tx.commit().map_err(|e| e.to_string())?;
        }
        "retry" => {
            if payload.is_some() || base_version.is_some() {
                if operation.status == "pending" {
                    return Err(
                        "Reconcile the original request before changing its contents".into(),
                    );
                }
                if operation.kind == "attachment.upload" {
                    return Err("Discard this upload and choose the file again to change it".into());
                }
                if let Some(value) = payload {
                    operation.payload = value;
                }
                if let Some(value) = base_version {
                    operation.base_version = Some(value);
                }
                operation.id = uuid::Uuid::new_v4().to_string();
                super::model::validate_operation(&operation)?;
                let mut db = runtime.store.0.lock().map_err(|e| e.to_string())?;
                let tx = db.transaction().map_err(|e| e.to_string())?;
                operation.status = "pending".into();
                operation.attempted = false;
                operation.error = None;
                operation.server_version = None;
                operation.server_value = None;
                tx.execute(
                    "UPDATE operations SET id=?,operation=? WHERE account=? AND workspace=? AND id=?",
                    params![
                        operation.id,
                        serde_json::to_string(&operation).map_err(|e| e.to_string())?,
                        account,
                        workspace_id,
                        operation_id,
                    ],
                )
                .map_err(|e| e.to_string())?;
                tx.commit().map_err(|e| e.to_string())?;
            } else {
                if operation.status == "conflict" {
                    return Err(
                        "Review the current server version before retrying this conflict".into(),
                    );
                }
                operation.status = "pending".into();
                operation.error = None;
                runtime.store.mark_operation(&account, &operation)?;
            }
        }
        _ => return Err("Unknown resolution".into()),
    }
    runtime.store.workspace(&account, &workspace_id)
}

#[tauri::command]
pub async fn desktop_logout(
    window: WebviewWindow,
    expected_account_id: String,
    discard_local_work: Option<bool>,
) -> Result<(), String> {
    let runtime = runtime(&window)?;
    let _guard = runtime.sync.lock().await;
    let account = super::require_account(&runtime, &expected_account_id)?;
    {
        let db = runtime.store.0.lock().map_err(|e| e.to_string())?;
        let pending:i64=db.query_row("SELECT (SELECT COUNT(*) FROM operations WHERE account=?) + (SELECT COUNT(*) FROM drafts WHERE account=? AND (kind<>'collaboration' OR json_extract(payload,'$.dirty')<>0))",params![account,account],|r|r.get(0)).map_err(|e|e.to_string())?;
        if pending > 0 && discard_local_work != Some(true) {
            return Err(
                "Save or export your drafts and resolve pending changes before signing out.".into(),
            );
        }
    }
    if let Ok(token) = runtime.transport.token(&account).await {
        let result = runtime
            .transport
            .request(reqwest::Method::DELETE, "/users/session", &token)?
            .send()
            .await;
        match result {
            Ok(response) if response.status().is_success() => {}
            Ok(response) => eprintln!(
                "Remote logout was not accepted: {}",
                response.status().as_u16()
            ),
            Err(_) => eprintln!("Remote logout could not reach the server"),
        }
    }
    runtime.transport.remove_session(&account).await?;
    {
        let mut db = runtime.store.0.lock().map_err(|e| e.to_string())?;
        let tx = db.transaction().map_err(|e| e.to_string())?;
        for table in [
            "snapshots",
            "drafts",
            "operations",
            "attachments",
            "workspace_settings",
            "workspace_access",
            "cached_files",
            "notification_receipts",
            "identity_aliases",
        ] {
            tx.execute(&format!("DELETE FROM {table} WHERE account=?"), [&account])
                .map_err(|e| e.to_string())?;
        }
        tx.execute("DELETE FROM accounts WHERE id=?", [&account])
            .map_err(|e| e.to_string())?;
        tx.execute(
            "DELETE FROM settings WHERE key IN ('account','workspace','pendingNavigation')",
            [],
        )
        .map_err(|e| e.to_string())?;
        tx.commit().map_err(|e| e.to_string())?;
    }
    window
        .emit("desktop-state-changed", ())
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub fn desktop_open_external(window: WebviewWindow, url: String) -> Result<(), String> {
    runtime(&window)?;
    let url = tauri::Url::parse(&url).map_err(|_| "Invalid URL")?;
    if !crate::navigation::can_open_externally(&url) {
        return Err("Unsupported link".into());
    }
    window
        .app_handle()
        .opener()
        .open_url(url.as_str(), None::<&str>)
        .map_err(|e| e.to_string())
}
