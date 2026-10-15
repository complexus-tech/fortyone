//! User-selected recovery exports acknowledge success only after durable write.
use super::*;
use tauri_plugin_dialog::DialogExt;

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct RecoveryRequest {
    account_id: String,
    workspace_id: String,
    document_id: String,
    payload: Value,
}

fn recovery(request: &RecoveryRequest) -> Result<Vec<u8>, String> {
    uuid::Uuid::parse_str(&request.document_id).map_err(|_| "Invalid document identifier")?;
    if !request.payload.is_object() {
        return Err("Invalid recovery document".into());
    }
    let bytes=serde_json::to_vec(&serde_json::json!({"format":"fortyone-document-recovery","version":1,"accountId":request.account_id,"workspaceId":request.workspace_id,"documentId":request.document_id,"exportedAt":now(),"payload":request.payload})).map_err(|error|error.to_string())?;
    if bytes.len() > 2 * 1024 * 1024 {
        return Err("The recovery document exceeds the 2 MiB export limit".into());
    }
    Ok(bytes)
}

pub(super) async fn write(path: std::path::PathBuf, bytes: Vec<u8>) -> Result<(), String> {
    tauri::async_runtime::spawn_blocking(move || atomic_write(&path, &bytes))
        .await
        .map_err(|error| error.to_string())?
}

fn atomic_write(path: &Path, bytes: &[u8]) -> Result<(), String> {
    use std::io::Write;
    let parent = path.parent().ok_or("Invalid destination")?;
    let temporary = parent.join(format!(".fortyone-{}.tmp", uuid::Uuid::new_v4()));
    let result = (|| {
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
        file.write_all(bytes)
            .map_err(|_| "Could not write the export. Check free disk space.")?;
        file.sync_all()
            .map_err(|_| "Could not save the export to disk")?;
        std::fs::rename(&temporary, path).map_err(|_| "Could not finish the export")?;
        Ok(())
    })();
    if result.is_err() {
        let _ = std::fs::remove_file(temporary);
    }
    result
}

#[tauri::command]
pub async fn hosted_export_recovery(
    window: WebviewWindow,
    request: RecoveryRequest,
) -> Result<bool, String> {
    let bytes = recovery(&request)?;
    let state = authorize_window(&window)?;
    let scope =
        super::drafts::authorize_known(&window, &request.account_id, &request.workspace_id).await?;
    let (send, receive) = tokio::sync::oneshot::channel();
    window
        .dialog()
        .file()
        .set_file_name(format!("FortyOne-document-{}.json", request.document_id))
        .save_file(move |path| {
            let _ = send.send(path);
        });
    let Some(path) = receive.await.map_err(|_| "Save dialog closed")? else {
        return Ok(false);
    };
    let refreshed =
        super::drafts::authorize_known(&window, &request.account_id, &request.workspace_id).await?;
    if scope.origin != refreshed.origin || scope.cookie_hash != refreshed.cookie_hash {
        return Err("Your account changed before exporting".into());
    }
    {
        let db = state.db.lock().map_err(|error| error.to_string())?;
        super::drafts::known_scope(&db, &refreshed)?;
    }
    write(
        path.into_path().map_err(|_| "Choose a local folder")?,
        bytes,
    )
    .await?;
    Ok(true)
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn recovery_has_bounded_identity_and_preserves_supplied_state() {
        let mut request = RecoveryRequest {
            account_id: "account".into(),
            workspace_id: "workspace".into(),
            document_id: uuid::Uuid::new_v4().to_string(),
            payload: serde_json::json!({"state":"base64","dirty":true}),
        };
        let value: Value = serde_json::from_slice(&recovery(&request).unwrap()).unwrap();
        assert_eq!(value["payload"], request.payload);
        assert_eq!(value["documentId"], request.document_id);
        request.payload = serde_json::json!({"state":"x".repeat(2*1024*1024)});
        assert!(recovery(&request).is_err());
    }
    #[test]
    fn completed_export_is_written_and_failed_replacement_keeps_destination() {
        let directory =
            std::env::temp_dir().join(format!("fortyone-export-{}", uuid::Uuid::new_v4()));
        std::fs::create_dir(&directory).unwrap();
        let path = directory.join("recovery.json");
        atomic_write(&path, b"retained").unwrap();
        assert_eq!(std::fs::read(&path).unwrap(), b"retained");
        let blocked = directory.join("blocked");
        std::fs::create_dir(&blocked).unwrap();
        assert!(atomic_write(&blocked, b"new").is_err());
        assert!(blocked.is_dir());
        assert_eq!(std::fs::read_dir(&directory).unwrap().count(), 2);
        std::fs::remove_dir_all(directory).unwrap();
    }
}
