use base64::Engine;
use rusqlite::params;
use serde_json::{json, Value};
use tauri::WebviewWindow;
use tauri_plugin_dialog::DialogExt;

use super::runtime;

pub async fn save(window: &WebviewWindow, name: &str, contents: Vec<u8>) -> Result<bool, String> {
    let (send, receive) = tokio::sync::oneshot::channel();
    window
        .dialog()
        .file()
        .set_file_name(name)
        .save_file(move |path| {
            let _ = send.send(path);
        });
    let Some(path) = receive.await.map_err(|_| "Save dialog closed")? else {
        return Ok(false);
    };
    let path = path.into_path().map_err(|_| "Choose a local folder")?;
    tauri::async_runtime::spawn_blocking(move || {
        use std::io::Write;
        // Write beside the destination and rename only after a durable flush;
        // a failed export never truncates an existing recovery file.
        let parent = path.parent().ok_or("Invalid destination")?;
        let temporary = parent.join(format!(".fortyone-{}.tmp", uuid::Uuid::new_v4()));
        let result = (|| -> Result<(), String> {
            let mut options = std::fs::OpenOptions::new();
            options.write(true).create_new(true);
            #[cfg(unix)]
            {
                use std::os::unix::fs::OpenOptionsExt;
                options.mode(0o600);
            }
            let mut file = options
                .open(&temporary)
                .map_err(|_| "Could not create the export file")?;
            file.write_all(&contents)
                .map_err(|_| "Could not write the export. Check free disk space.")?;
            file.sync_all()
                .map_err(|_| "Could not save the export to disk")?;
            std::fs::rename(&temporary, &path).map_err(|_| "Could not finish the export")?;
            Ok(())
        })();
        if result.is_err() {
            let _ = std::fs::remove_file(temporary);
        }
        result
    })
    .await
    .map_err(|e| e.to_string())??;
    Ok(true)
}

#[tauri::command]
pub async fn desktop_export_document(
    window: WebviewWindow,
    title: String,
    content_html: String,
    content_text: String,
) -> Result<bool, String> {
    runtime(&window)?;
    if title.len() > 1024 || content_text.len() + content_html.len() > 16 * 1024 * 1024 {
        return Err("The document exceeds the export limit".into());
    }
    // Plain text is portable and cannot execute scripts copied from a document.
    let name: String = title
        .chars()
        .filter(|c| c.is_alphanumeric() || matches!(c, ' ' | '-' | '_'))
        .take(80)
        .collect();
    let name = if name.trim().is_empty() {
        "Document"
    } else {
        name.trim()
    };
    save(
        &window,
        &format!("{name}.txt"),
        format!("{title}\n\n{content_text}").into_bytes(),
    )
    .await
}

#[tauri::command]
pub async fn desktop_export_local_work(
    window: WebviewWindow,
    expected_account_id: String,
) -> Result<bool, String> {
    let runtime = runtime(&window)?;
    let account = super::require_account(&runtime, &expected_account_id)?;
    let export = {
        let db = runtime.store.0.lock().map_err(|e| e.to_string())?;
        let mut result = json!({"format":"fortyone-recovery","version":1,"accountId":account,"exportedAt":super::model::now()});
        for (table, column) in [("operations", "operation"), ("snapshots", "data")] {
            let mut query = db
                .prepare(&format!(
                    "SELECT workspace,{column} FROM {table} WHERE account=?"
                ))
                .map_err(|e| e.to_string())?;
            let rows = query
                .query_map([&account], |r| {
                    Ok((r.get::<_, String>(0)?, r.get::<_, String>(1)?))
                })
                .map_err(|e| e.to_string())?;
            let mut values = Vec::new();
            for row in rows {
                let (workspace, data) = row.map_err(|e| e.to_string())?;
                values.push(json!({"workspaceId":workspace,"data":serde_json::from_str::<Value>(&data).map_err(|e|e.to_string())?}));
            }
            result[table] = json!(values);
        }
        let mut drafts = db
            .prepare("SELECT workspace,resource,kind,payload FROM drafts WHERE account=?")
            .map_err(|e| e.to_string())?;
        let values = drafts
            .query_map([&account], |r| {
                Ok((
                    r.get::<_, String>(0)?,
                    r.get::<_, String>(1)?,
                    r.get::<_, String>(2)?,
                    r.get::<_, String>(3)?,
                ))
            })
            .map_err(|e| e.to_string())?;
        let mut draft_values = Vec::new();
        for row in values {
            let (workspace, resource, kind, payload) = row.map_err(|e| e.to_string())?;
            draft_values.push(json!({"workspaceId":workspace,"resourceId":resource,"kind":kind,"payload":serde_json::from_str::<Value>(&payload).map_err(|e|e.to_string())?}));
        }
        result["drafts"] = json!(draft_values);
        let mut query = db
            .prepare("SELECT id,workspace,name,bytes FROM attachments WHERE account=?")
            .map_err(|e| e.to_string())?;
        let rows = query
            .query_map(params![account], |r| {
                Ok((
                    r.get::<_, String>(0)?,
                    r.get::<_, String>(1)?,
                    r.get::<_, String>(2)?,
                    r.get::<_, Vec<u8>>(3)?,
                ))
            })
            .map_err(|e| e.to_string())?;
        let mut files = Vec::new();
        for row in rows {
            let (id, workspace, name, bytes) = row.map_err(|e| e.to_string())?;
            files.push(json!({"id":id,"workspaceId":workspace,"name":name,"data":base64::engine::general_purpose::STANDARD.encode(bytes)}));
        }
        result["attachments"] = json!(files);
        serde_json::to_vec_pretty(&result).map_err(|e| e.to_string())?
    };
    save(&window, "FortyOne local work.json", export).await
}
