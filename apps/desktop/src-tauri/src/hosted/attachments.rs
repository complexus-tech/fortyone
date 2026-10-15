//! Retained attachment bytes never enter the outbox status payload. The upload
//! receipt identifies the file independently from the story that owns it.
use super::*;
use base64::Engine;
use serde_json::json;
use tauri_plugin_dialog::DialogExt;

const MAX_BYTES: usize = 10 * 1024 * 1024;
const MAX_STAGED_BYTES: usize = 128 * 1024 * 1024;

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct StageRequest {
    account_id: String,
    workspace_id: String,
    id: String,
    story_id: String,
    file_name: String,
    mime_type: String,
    data: String,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ReadRequest {
    account_id: String,
    workspace_id: String,
    story_id: String,
    attachment_id: String,
}

pub(super) fn initialize(db: &Connection) -> Result<(), String> {
    db.execute_batch("CREATE TABLE IF NOT EXISTS hosted_files(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,id TEXT NOT NULL,server_id TEXT,story TEXT NOT NULL,name TEXT NOT NULL,mime TEXT NOT NULL,bytes BLOB NOT NULL,hash TEXT NOT NULL,created INTEGER NOT NULL,uploaded INTEGER NOT NULL DEFAULT 0,PRIMARY KEY(origin,account,workspace,id));").map_err(|error| error.to_string())
}

fn validate(request: &StageRequest) -> Result<Vec<u8>, String> {
    uuid::Uuid::parse_str(&request.id).map_err(|_| "Invalid attachment identifier")?;
    uuid::Uuid::parse_str(&request.story_id).map_err(|_| "Invalid story identifier")?;
    if request.file_name.trim().is_empty()
        || request.file_name.len() > 255
        || request
            .file_name
            .chars()
            .any(|c| c.is_control() || matches!(c, '/' | '\\'))
    {
        return Err("Choose a file with a valid filename".into());
    }
    let mime = request.mime_type.split(';').next().unwrap_or_default();
    if mime.len() > 128
        || mime.split('/').count() != 2
        || mime.split('/').any(str::is_empty)
        || !mime
            .bytes()
            .all(|c| c.is_ascii_alphanumeric() || b"/!#$&^_.+-".contains(&c))
        || request.mime_type.len() != mime.len()
    {
        return Err("The attachment has an invalid media type".into());
    }
    if request.data.len() > MAX_BYTES.div_ceil(3) * 4 {
        return Err("Desktop attachments must be 10 MiB or smaller".into());
    }
    let bytes = base64::engine::general_purpose::STANDARD
        .decode(&request.data)
        .map_err(|_| "The attachment data is invalid")?;
    if bytes.is_empty() || bytes.len() > MAX_BYTES {
        return Err("Choose a nonempty attachment of 10 MiB or smaller".into());
    }
    Ok(bytes)
}

fn stage(
    db: &Connection,
    scope: &BoundScope,
    request: &StageRequest,
    bytes: &[u8],
) -> Result<Value, String> {
    let hash = format!("{:x}", Sha256::digest(bytes));
    let previous: Option<(String,String,String)> = db.query_row("SELECT name,mime,hash FROM hosted_files WHERE origin=? AND account=? AND workspace=? AND id=?", params![scope.origin,scope.account,scope.workspace,request.id], |row| Ok((row.get(0)?,row.get(1)?,row.get(2)?))).optional().map_err(|error|error.to_string())?;
    if let Some((name, mime, previous_hash)) = previous {
        if name != request.file_name || mime != request.mime_type || previous_hash != hash {
            return Err("This attachment ID was already used for a different file".into());
        }
    } else {
        let mut size:usize = db.query_row("SELECT COALESCE(SUM(LENGTH(bytes)),0) FROM hosted_files WHERE origin=? AND account=?",params![scope.origin,scope.account],|row|row.get(0)).map_err(|error|error.to_string())?;
        while size.saturating_add(bytes.len()) > MAX_STAGED_BYTES {
            let oldest:Option<(i64,usize)>=db.query_row("SELECT rowid,LENGTH(bytes) FROM hosted_files WHERE origin=? AND account=? AND uploaded=1 ORDER BY created LIMIT 1",params![scope.origin,scope.account],|row|Ok((row.get(0)?,row.get(1)?))).optional().map_err(|error|error.to_string())?;
            let Some((row, removed)) = oldest else { break };
            db.execute("DELETE FROM hosted_files WHERE rowid=?", [row])
                .map_err(|error| error.to_string())?;
            size = size.saturating_sub(removed);
        }
        if size.saturating_add(bytes.len()) > MAX_STAGED_BYTES {
            return Err(
                "The local attachment store is full. Synchronize pending files before adding more."
                    .into(),
            );
        }
        db.execute("INSERT INTO hosted_files(origin,account,workspace,id,story,name,mime,bytes,hash,created) VALUES(?,?,?,?,?,?,?,?,?,?)",params![scope.origin,scope.account,scope.workspace,request.id,request.story_id,request.file_name,request.mime_type,bytes,hash,now()]).map_err(|_| "Could not retain this attachment. Check available disk space.".to_owned())?;
    }
    let created:String=db.query_row("SELECT strftime('%Y-%m-%dT%H:%M:%fZ',created/1000.0,'unixepoch') FROM hosted_files WHERE origin=? AND account=? AND workspace=? AND id=?",params![scope.origin,scope.account,scope.workspace,request.id],|row|row.get(0)).map_err(|error|error.to_string())?;
    let operation = outbox::EnqueueRequest {
        account_id: scope.account.clone(),
        workspace_id: scope.workspace.clone(),
        id: request.id.clone(),
        kind: "attachment.upload".into(),
        resource_id: request.story_id.clone(),
        payload: json!({"attachmentId":request.id,"targetKind":"story","fileName":request.file_name,"contentType":request.mime_type,"contentHash":hash,"size":bytes.len(),"createdAt":created}),
        base_version: None,
        consumed_draft_key: None,
    };
    let result = outbox::store::enqueue(db, scope, &operation)?;
    // A story creation may already have resolved before staging or while this
    // same input is being recovered after an interrupted IPC response.
    if result.pending {
        db.execute(
            "UPDATE hosted_files SET story=? WHERE origin=? AND account=? AND workspace=? AND id=?",
            params![
                result.resource_id,
                scope.origin,
                scope.account,
                scope.workspace,
                request.id
            ],
        )
        .map_err(|error| error.to_string())?;
    }
    if !result.pending {
        let story = outbox::store::resolve_resource(db, scope, &request.story_id)?;
        db.execute("UPDATE hosted_files SET story=?,server_id=?,uploaded=1 WHERE origin=? AND account=? AND workspace=? AND id=?",params![story,result.resource_id,scope.origin,scope.account,scope.workspace,request.id]).map_err(|error|error.to_string())?;
    }
    let attachment:Option<String>=db.query_row("SELECT server_id FROM hosted_files WHERE origin=? AND account=? AND workspace=? AND id=?",params![scope.origin,scope.account,scope.workspace,request.id],|row|row.get(0)).map_err(|error|error.to_string())?;
    Ok(
        json!({"operationId":result.operation_id,"attachment":{"id":attachment.unwrap_or_else(||request.id.clone()),"filename":request.file_name,"size":bytes.len(),"mimeType":request.mime_type,"url":"","createdAt":created,"uploadedBy":scope.account,"pending":result.pending}}),
    )
}

#[tauri::command]
pub async fn hosted_stage_attachment(
    window: WebviewWindow,
    request: StageRequest,
) -> Result<Value, String> {
    let bytes = validate(&request)?;
    let state = authorize_window(&window)?;
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let mut db = state.db.lock().map_err(|error| error.to_string())?;
    let tx = db.transaction().map_err(|error| error.to_string())?;
    ensure_scope_current(&tx, &scope)?;
    let result = stage(&tx, &scope, &request, &bytes)?;
    tx.commit().map_err(|error| error.to_string())?;
    Ok(result)
}

pub(super) fn wire_payload(
    db: &Connection,
    scope: &BoundScope,
    operation: &crate::operations::Operation,
) -> Result<Value, String> {
    let id = operation.payload["attachmentId"]
        .as_str()
        .ok_or("Missing retained attachment")?;
    let file:Option<(String,String,Vec<u8>,String)>=db.query_row("SELECT name,mime,bytes,hash FROM hosted_files WHERE origin=? AND account=? AND workspace=? AND id=? AND story=? AND uploaded=0",params![scope.origin,scope.account,scope.workspace,id,operation.resource_id],|row|Ok((row.get(0)?,row.get(1)?,row.get(2)?,row.get(3)?))).optional().map_err(|error|error.to_string())?;
    let (name, mime, bytes, hash) =
        file.ok_or("The locally saved attachment is missing. It has not been uploaded.")?;
    if operation.payload["contentHash"].as_str() != Some(&hash)
        || bytes.len() > MAX_BYTES
        || format!("{:x}", Sha256::digest(&bytes)) != hash
    {
        return Err("The retained attachment no longer matches this upload".into());
    }
    Ok(
        json!({"fileName":name,"contentType":mime,"targetKind":"story","data":base64::engine::general_purpose::STANDARD.encode(bytes)}),
    )
}

fn read(
    db: &Connection,
    scope: &BoundScope,
    request: &ReadRequest,
) -> Result<Option<(String, String, Vec<u8>)>, String> {
    uuid::Uuid::parse_str(&request.story_id).map_err(|_| "Invalid story identifier")?;
    uuid::Uuid::parse_str(&request.attachment_id).map_err(|_| "Invalid attachment identifier")?;
    let story = outbox::store::resolve_resource(db, scope, &request.story_id)?;
    db.query_row("SELECT name,mime,bytes FROM hosted_files WHERE origin=? AND account=? AND workspace=? AND story=? AND (id=? OR server_id=?)",params![scope.origin,scope.account,scope.workspace,story,request.attachment_id,request.attachment_id],|row|Ok((row.get(0)?,row.get(1)?,row.get(2)?))).optional().map_err(|error|error.to_string())
}

#[tauri::command]
pub async fn hosted_read_attachment(
    window: WebviewWindow,
    request: ReadRequest,
) -> Result<Option<Value>, String> {
    let state = authorize_window(&window)?;
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let db = state.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, &scope)?;
    Ok(read(&db,&scope,&request)?.map(|(file_name,mime_type,bytes)|json!({"fileName":file_name,"mimeType":mime_type,"data":base64::engine::general_purpose::STANDARD.encode(bytes)})))
}

#[tauri::command]
pub async fn hosted_export_attachment(
    window: WebviewWindow,
    request: ReadRequest,
) -> Result<bool, String> {
    let state = authorize_window(&window)?;
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let name = {
        let db = state.db.lock().map_err(|error| error.to_string())?;
        ensure_scope_current(&db, &scope)?;
        read(&db, &scope, &request)?
            .ok_or("This file is not saved on this device")?
            .0
    };
    let (send, receive) = tokio::sync::oneshot::channel();
    window
        .dialog()
        .file()
        .set_file_name(&name)
        .save_file(move |path| {
            let _ = send.send(path);
        });
    let Some(path) = receive.await.map_err(|_| "Save dialog closed")? else {
        return Ok(false);
    };
    let path = path.into_path().map_err(|_| "Choose a local folder")?;
    let refreshed =
        authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    if refreshed.cookie_hash != scope.cookie_hash || refreshed.origin != scope.origin {
        return Err("Your account changed before exporting".into());
    }
    let bytes = {
        let db = state.db.lock().map_err(|error| error.to_string())?;
        ensure_scope_current(&db, &refreshed)?;
        read(&db, &refreshed, &request)?
            .ok_or("This file is no longer saved on this device")?
            .2
    };
    super::export::write(path, bytes).await?;
    Ok(true)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::operations::Operation;

    fn database() -> Connection {
        let db = Connection::open_in_memory().unwrap();
        super::super::initialize(&db).unwrap();
        db
    }
    fn request(scope: &BoundScope, story: &str) -> StageRequest {
        StageRequest {
            account_id: scope.account.clone(),
            workspace_id: scope.workspace.clone(),
            id: uuid::Uuid::new_v4().to_string(),
            story_id: story.into(),
            file_name: "report.pdf".into(),
            mime_type: "application/pdf".into(),
            data: base64::engine::general_purpose::STANDARD.encode(b"retained file"),
        }
    }
    fn staged(db: &mut Connection, scope: &BoundScope, request: &StageRequest) -> Value {
        let tx = db.transaction().unwrap();
        let result = stage(&tx, scope, request, &validate(request).unwrap()).unwrap();
        tx.commit().unwrap();
        result
    }
    fn acknowledge(db: &mut Connection, scope: &BoundScope, operation: &Operation, result: Value) {
        let tx = db.transaction().unwrap();
        outbox::store::acknowledge(&tx, scope, operation, &result).unwrap();
        tx.commit().unwrap();
    }
    fn file(scope: &BoundScope, story: &str, id: &str) -> ReadRequest {
        ReadRequest {
            account_id: scope.account.clone(),
            workspace_id: scope.workspace.clone(),
            story_id: story.into(),
            attachment_id: id.into(),
        }
    }

    #[test]
    fn upload_ack_preserves_story_identity_and_retains_exact_file_binding() {
        let mut db = database();
        let scope = super::super::tests::scope();
        let story = uuid::Uuid::new_v4().to_string();
        let request = request(&scope, &story);
        let first = staged(&mut db, &scope, &request);
        assert!(first["attachment"]["createdAt"]
            .as_str()
            .unwrap()
            .ends_with('Z'));
        assert_eq!(staged(&mut db, &scope, &request), first);
        let operation = outbox::store::read(&db, &scope)
            .unwrap()
            .operations
            .remove(0);
        assert!(operation.payload.get("data").is_none());
        assert_eq!(
            wire_payload(&db, &scope, &operation).unwrap()["data"],
            request.data
        );
        let next = outbox::EnqueueRequest {
            account_id: scope.account.clone(),
            workspace_id: scope.workspace.clone(),
            id: uuid::Uuid::new_v4().to_string(),
            kind: "story.update".into(),
            resource_id: story.clone(),
            payload: json!({"title":"After upload"}),
            base_version: Some("v1".into()),
            consumed_draft_key: None,
        };
        outbox::store::enqueue(&db, &scope, &next).unwrap();
        let server = uuid::Uuid::new_v4().to_string();
        acknowledge(&mut db, &scope, &operation, json!({"resourceId":server}));
        let pending = outbox::store::read(&db, &scope).unwrap();
        assert_eq!(pending.operations[0].resource_id, story);
        assert_eq!(pending.operations[0].base_version.as_deref(), Some("v1"));
        assert_eq!(pending.completed[0]["provisionalId"], request.id);
        assert_eq!(pending.completed[0]["storyId"], story);
        assert_eq!(pending.completed[0]["resource"]["id"], server);
        assert!(read(&db, &scope, &file(&scope, &story, &server))
            .unwrap()
            .is_some());
        assert!(read(
            &db,
            &scope,
            &file(&scope, &uuid::Uuid::new_v4().to_string(), &server)
        )
        .unwrap()
        .is_none());
        let replay = staged(&mut db, &scope, &request);
        assert_eq!(replay["attachment"]["id"], server);
        assert_eq!(replay["attachment"]["pending"], false);
        assert_eq!(
            outbox::store::read(&db, &scope).unwrap().operations.len(),
            1
        );
    }

    #[test]
    fn creation_ack_remaps_file_target_and_keeps_blob_out_of_operation_status() {
        let mut db = database();
        let scope = super::super::tests::scope();
        let local = uuid::Uuid::new_v4().to_string();
        let create = outbox::EnqueueRequest {
            account_id: scope.account.clone(),
            workspace_id: scope.workspace.clone(),
            id: uuid::Uuid::new_v4().to_string(),
            kind: "story.create".into(),
            resource_id: local.clone(),
            payload: json!({"title":"Local"}),
            base_version: None,
            consumed_draft_key: None,
        };
        outbox::store::enqueue(&db, &scope, &create).unwrap();
        let request = request(&scope, &local);
        staged(&mut db, &scope, &request);
        let operation = outbox::store::read(&db, &scope)
            .unwrap()
            .operations
            .remove(0);
        let remote = uuid::Uuid::new_v4().to_string();
        acknowledge(
            &mut db,
            &scope,
            &operation,
            json!({"resourceId":remote,"operationVersion":"v1"}),
        );
        let upload = outbox::store::read(&db, &scope)
            .unwrap()
            .operations
            .remove(0);
        assert_eq!(upload.resource_id, remote);
        assert!(wire_payload(&db, &scope, &upload).is_ok());
        assert!(read(&db, &scope, &file(&scope, &local, &request.id))
            .unwrap()
            .is_some());
        let mut other = super::super::tests::scope();
        other.account = uuid::Uuid::new_v4().to_string();
        assert!(read(&db, &other, &file(&other, &remote, &request.id))
            .unwrap()
            .is_none());
    }

    #[test]
    fn staging_queue_failure_rolls_back_bytes_and_reused_ids_reject_changed_files() {
        let mut db = database();
        let scope = super::super::tests::scope();
        let story = uuid::Uuid::new_v4().to_string();
        let mut request = request(&scope, &story);
        db.execute_batch("CREATE TRIGGER reject_upload BEFORE INSERT ON hosted_operations BEGIN SELECT RAISE(ABORT,'full'); END;").unwrap();
        {
            let tx = db.transaction().unwrap();
            assert!(stage(&tx, &scope, &request, &validate(&request).unwrap()).is_err());
        }
        assert_eq!(
            db.query_row("SELECT COUNT(*) FROM hosted_files", [], |row| row
                .get::<_, i64>(0))
                .unwrap(),
            0
        );
        db.execute_batch("DROP TRIGGER reject_upload").unwrap();
        staged(&mut db, &scope, &request);
        request.data = base64::engine::general_purpose::STANDARD.encode(b"different");
        assert!(stage(&db, &scope, &request, &validate(&request).unwrap()).is_err());
        assert_eq!(
            outbox::store::read(&db, &scope).unwrap().operations.len(),
            1
        );
    }

    #[test]
    fn uploaded_cache_is_purged_on_revocation_but_unsent_bytes_survive() {
        let mut db = database();
        let scope = super::super::tests::scope();
        let story = uuid::Uuid::new_v4().to_string();
        let first = request(&scope, &story);
        staged(&mut db, &scope, &first);
        let operation = outbox::store::read(&db, &scope)
            .unwrap()
            .operations
            .remove(0);
        acknowledge(
            &mut db,
            &scope,
            &operation,
            json!({"resourceId":uuid::Uuid::new_v4().to_string()}),
        );
        let pending = request(&scope, &story);
        staged(&mut db, &scope, &pending);
        super::super::forget_origin(&db, &scope.origin).unwrap();
        assert!(read(&db, &scope, &file(&scope, &story, &first.id))
            .unwrap()
            .is_none());
        assert!(read(&db, &scope, &file(&scope, &story, &pending.id))
            .unwrap()
            .is_some());
        assert!(super::super::ensure_scope_current(&db, &scope).is_err());
    }

    #[test]
    fn validates_filename_mime_and_size_before_retaining_data() {
        let scope = super::super::tests::scope();
        let mut request = request(&scope, &uuid::Uuid::new_v4().to_string());
        request.file_name = "../file.pdf".into();
        assert!(validate(&request).is_err());
        request.file_name = "file.pdf".into();
        request.mime_type = "text/html\r\nheader:1".into();
        assert!(validate(&request).is_err());
        request.mime_type = "application/pdf".into();
        request.data = "a".repeat(MAX_BYTES.div_ceil(3) * 4 + 1);
        assert!(validate(&request).is_err());
    }
}
