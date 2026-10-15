use super::{model::Operation, runtime};
use rusqlite::params;
use serde_json::{json, Value};
use tauri::WebviewWindow;
use tauri_plugin_dialog::DialogExt;

#[tauri::command]
pub async fn desktop_stage_attachment(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    resource_id: String,
) -> Result<Value, String> {
    let runtime = runtime(&window)?;
    let account = super::require_account(&runtime, &expected_account_id)?;
    runtime.store.authorize(&account, &workspace_id)?;
    uuid::Uuid::parse_str(&resource_id).map_err(|_| "Choose a story before attaching a file")?;
    let (send, receive) = tokio::sync::oneshot::channel();
    window.dialog().file().pick_file(move |file| {
        let _ = send.send(file);
    });
    let Some(file) = receive.await.map_err(|_| "File picker closed")? else {
        return runtime.store.workspace(&account, &workspace_id);
    };
    let path = file.into_path().map_err(|_| "Choose a local file")?;
    let bytes = tauri::async_runtime::spawn_blocking(move || {
        let metadata = std::fs::metadata(&path).map_err(|_| "Could not read the selected file")?;
        if !metadata.is_file() || metadata.len() > 10 * 1024 * 1024 {
            return Err("Choose a file smaller than 10 MB");
        }
        let name = path
            .file_name()
            .and_then(|v| v.to_str())
            .ok_or("Invalid file name")?
            .to_owned();
        // Read a bounded amount even if the file changes after metadata lookup.
        use std::io::Read;
        let mut data = Vec::new();
        std::fs::File::open(path)
            .map_err(|_| "Could not open the file")?
            .take(10 * 1024 * 1024 + 1)
            .read_to_end(&mut data)
            .map_err(|_| "Could not read the file")?;
        if data.len() > 10 * 1024 * 1024 {
            return Err("The selected file grew beyond the attachment limit");
        }
        Ok((name, data))
    })
    .await
    .map_err(|e| e.to_string())??;
    if runtime.store.account()? != account {
        return Err("Your account changed. Choose the attachment again.".into());
    }
    runtime.store.authorize(&account, &workspace_id)?;
    let id = uuid::Uuid::new_v4().to_string();
    let mut operation = Operation {
        id: id.clone(),
        kind: "attachment.upload".into(),
        resource_id,
        payload: json!({"attachmentId":id,"fileName":bytes.0,"size":bytes.1.len(),"targetKind":"story"}),
        base_version: None,
        status: "pending".into(),
        error: None,
        server_version: None,
        server_value: None,
        attempted: false,
    };
    {
        let mut db = runtime.store.0.lock().map_err(|e| e.to_string())?;
        let tx = db.transaction().map_err(|e| e.to_string())?;
        super::require_account_in(&tx, &account)?;
        operation.resource_id = super::identity::resolve_resource(
            &tx,
            &account,
            &workspace_id,
            &operation.resource_id,
        )?;
        tx.execute(
            "INSERT INTO attachments VALUES(?,?,?,?,?)",
            params![id, account, workspace_id, bytes.0, bytes.1],
        )
        .map_err(|_| "Could not stage this file. Check free disk space.".to_owned())?;
        tx.execute(
            "INSERT INTO operations VALUES(?,?,?,?)",
            params![
                id,
                account,
                workspace_id,
                serde_json::to_string(&operation).map_err(|e| e.to_string())?
            ],
        )
        .map_err(|e| e.to_string())?;
        tx.commit().map_err(|e| e.to_string())?;
    }
    runtime.store.workspace(&account, &workspace_id)
}
