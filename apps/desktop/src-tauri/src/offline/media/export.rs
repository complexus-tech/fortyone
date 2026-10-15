use super::{blob_url, fetch_blob, runtime, verify_scope, Scope, DOWNLOADS, MAX_MEDIA_BYTES};
use rusqlite::{params, OptionalExtension};
use tauri::WebviewWindow;

fn file_name(value: &str, fallback: &str) -> String {
    let value: String = value
        .chars()
        .filter(|character| {
            !character.is_control()
                && !matches!(
                    character,
                    '/' | '\\' | ':' | '*' | '?' | '"' | '<' | '>' | '|'
                )
        })
        .take(180)
        .collect();
    let value = value.trim().trim_matches('.');
    if value.is_empty() {
        fallback.into()
    } else {
        value.into()
    }
}

fn local_file(store: &super::Store, scope: &Scope) -> Result<Option<(String, Vec<u8>)>, String> {
    let db = store.0.lock().map_err(|error| error.to_string())?;
    verify_scope(&db, scope)?;
    // Staged and retained uploads have no resource column. Verify their exact
    // binding through the durable outbox or downloaded attachment metadata.
    let retained:Option<(String,Vec<u8>)>=db.query_row(
        "SELECT name,bytes FROM attachments WHERE id=?1 AND account=?2 AND workspace=?3 AND (EXISTS(SELECT 1 FROM operations WHERE account=?2 AND workspace=?3 AND json_extract(operation,'$.resourceId')=?4 AND json_extract(operation,'$.payload.targetKind')=?5 AND json_extract(operation,'$.payload.attachmentId')=?1) OR EXISTS(SELECT 1 FROM snapshots,json_each(snapshots.data,'$.attachments') WHERE snapshots.account=?2 AND snapshots.workspace=?3 AND json_extract(value,'$.id')=?1 AND json_extract(value,'$.resourceId')=?4 AND json_extract(value,'$.targetKind')=?5))",
        params![scope.attachment,scope.account,scope.workspace,scope.resource,scope.kind],|row|Ok((row.get(0)?,row.get(1)?))
    ).optional().map_err(|error|error.to_string())?;
    if retained.is_some() {
        return Ok(retained);
    }
    db.query_row("SELECT name,bytes FROM cached_files WHERE account=? AND workspace=? AND resource=? AND kind=? AND attachment=?",params![scope.account,scope.workspace,scope.resource,scope.kind,scope.attachment],|row|Ok((row.get(0)?,row.get(1)?))).optional().map_err(|error|error.to_string())
}

async fn remote_file(runtime: &super::Runtime, scope: &Scope) -> Result<(String, Vec<u8>), String> {
    let _permit = DOWNLOADS
        .acquire()
        .await
        .map_err(|_| "The attachment downloader is unavailable")?;
    if runtime.store.account()? != scope.account {
        return Err("Your account changed before the attachment download".into());
    }
    runtime.store.authorize(&scope.account, &scope.workspace)?;
    let token = runtime.transport.token(&scope.account).await?;
    let (name, url) = if scope.kind == "story" {
        // The story read enforces resource/team access before asking the legacy
        // attachment-list API for a freshly signed URL for this exact file.
        runtime
            .transport
            .get(
                &format!("/workspaces/{}/stories/{}", scope.slug, scope.resource),
                &token,
            )
            .await?;
        let files = runtime
            .transport
            .get(
                &format!(
                    "/workspaces/{}/stories/{}/attachments",
                    scope.slug, scope.resource
                ),
                &token,
            )
            .await?;
        let file = files
            .as_array()
            .and_then(|files| {
                files
                    .iter()
                    .find(|file| file["id"].as_str() == Some(&scope.attachment))
            })
            .ok_or("This attachment is no longer available")?;
        if file["size"]
            .as_u64()
            .is_some_and(|size| size > MAX_MEDIA_BYTES as u64)
        {
            return Err("Only attachments up to 10 MB can be exported from desktop".into());
        }
        let url = blob_url(
            file["url"]
                .as_str()
                .ok_or("Missing attachment download address")?,
        )?;
        (
            file_name(
                file["filename"].as_str().unwrap_or("Attachment"),
                "Attachment",
            ),
            url,
        )
    } else {
        let response = runtime
            .transport
            .request(reqwest::Method::GET, &scope.path(), &token)?
            .send()
            .await
            .map_err(|_| "Connect to download this attachment")?;
        if !matches!(response.status().as_u16(), 302 | 303 | 307 | 308) {
            return Err("This document attachment is no longer available".into());
        }
        let url = blob_url(
            response
                .headers()
                .get(reqwest::header::LOCATION)
                .and_then(|value| value.to_str().ok())
                .ok_or("Missing attachment download address")?,
        )?;
        ("Attachment".into(), url)
    };
    let (mime, bytes) = fetch_blob(url).await?;
    let name = if name == "Attachment" {
        let extension = match mime.as_str() {
            "image/png" => "png",
            "image/jpeg" => "jpg",
            "image/gif" => "gif",
            "image/webp" => "webp",
            "video/mp4" => "mp4",
            _ => "bin",
        };
        format!("Attachment.{extension}")
    } else {
        name
    };
    if runtime.store.account()? != scope.account {
        return Err("Your account changed during the attachment download".into());
    }
    runtime.store.authorize(&scope.account, &scope.workspace)?;
    Ok((name, bytes))
}

pub(super) async fn export_attachment(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    resource_id: String,
    kind: String,
    attachment_id: String,
) -> Result<bool, String> {
    let runtime = runtime(&window)?;
    let scope = Scope::new(
        &runtime,
        &expected_account_id,
        workspace_id,
        resource_id,
        kind,
        attachment_id,
    )?;
    let file = local_file(&runtime.store, &scope)?;
    let (name, bytes) = match file {
        Some(file) => file,
        None => remote_file(&runtime, &scope).await?,
    };
    if bytes.len() > MAX_MEDIA_BYTES {
        return Err("Only attachments up to 10 MB can be exported from desktop".into());
    }
    if runtime.store.account()? != scope.account {
        return Err("Your account changed. Reopen this attachment.".into());
    }
    runtime.store.authorize(&scope.account, &scope.workspace)?;
    {
        let db = runtime.store.0.lock().map_err(|error| error.to_string())?;
        verify_scope(&db, &scope)?;
    }
    super::super::recovery::save(&window, &file_name(&name, "Attachment"), bytes).await
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn export_names_never_supply_a_path() {
        assert_eq!(
            file_name("../../budget/report.pdf", "Attachment"),
            "budgetreport.pdf"
        );
        assert_eq!(file_name("..", "Attachment"), "Attachment");
        assert_eq!(file_name("report\n.pdf", "Attachment"), "report.pdf");
    }
    #[test]
    fn retained_uploads_require_the_exact_resource_binding() {
        use serde_json::json;
        let path =
            std::env::temp_dir().join(format!("fortyone-export-{}.sqlite", uuid::Uuid::new_v4()));
        let store = super::super::Store::open(&path).unwrap();
        super::super::initialize(&store).unwrap();
        store
            .save_account(&json!({"id":"account"}), &json!([{"id":"workspace"}]))
            .unwrap();
        let mut scope = Scope {
            account: "account".into(),
            workspace: "workspace".into(),
            resource: "story".into(),
            kind: "story".into(),
            attachment: "attachment".into(),
            slug: "acme".into(),
        };
        {
            let db = store.0.lock().unwrap();
            db.execute(
                "INSERT INTO attachments VALUES(?,?,?,?,?)",
                params![
                    "attachment",
                    "account",
                    "workspace",
                    "file.pdf",
                    b"pdf".as_slice()
                ],
            )
            .unwrap();
            db.execute("INSERT INTO operations VALUES(?,?,?,?)",params!["operation","account","workspace",json!({"kind":"attachment.upload","resourceId":"story","payload":{"targetKind":"story","attachmentId":"attachment"}}).to_string()]).unwrap();
            db.execute(
                "INSERT INTO operations VALUES(?,?,?,?)",
                params![
                    "creation",
                    "account",
                    "workspace",
                    json!({"kind":"story.create","resourceId":"story"}).to_string()
                ],
            )
            .unwrap();
        }
        assert!(local_file(&store, &scope).unwrap().is_some());
        store
            .0
            .lock()
            .unwrap()
            .execute(
                "INSERT INTO operations VALUES(?,?,?,?)",
                params![
                    "other-creation",
                    "account",
                    "workspace",
                    json!({"kind":"story.create","resourceId":"accessible-story"}).to_string()
                ],
            )
            .unwrap();
        scope.resource = "accessible-story".into();
        assert!(local_file(&store, &scope).unwrap().is_none());
        scope.resource = "another-story".into();
        assert!(local_file(&store, &scope).is_err());
        drop(store);
        for suffix in ["", "-wal", "-shm"] {
            let _ = std::fs::remove_file(format!("{}{suffix}", path.display()));
        }
    }
}
