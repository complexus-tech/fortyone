use super::*;

const MAX_DRAFT_BYTES: usize = 2 * 1024 * 1024;

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct DraftRequest {
    account_id: String,
    workspace_id: String,
    key: String,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct WriteRequest {
    account_id: String,
    workspace_id: String,
    key: String,
    payload: Value,
}

pub(super) fn validate_key(key: &str) -> Result<(), String> {
    let parts: Vec<&str> = key.split(':').collect();
    let valid = match parts.as_slice() {
        [kind, id]
            if matches!(
                *kind,
                "new-story" | "story" | "comment" | "document" | "collaboration"
            ) =>
        {
            uuid::Uuid::parse_str(id).is_ok()
        }
        ["comment", story, action, comment] if matches!(*action, "reply" | "edit") => {
            uuid::Uuid::parse_str(story).is_ok() && uuid::Uuid::parse_str(comment).is_ok()
        }
        _ => false,
    };
    if !valid {
        return Err("Invalid native draft key".into());
    }
    Ok(())
}

pub(super) fn known_scope(db: &Connection, scope: &BoundScope) -> Result<(), String> {
    let valid: bool = db.query_row(
        "SELECT EXISTS(SELECT 1 FROM hosted_known WHERE origin=?1 AND account=?2 AND workspace=?3 AND slug=?4 AND cookie_hash=?5) AND NOT EXISTS(SELECT 1 FROM hosted_active WHERE origin=?1 AND (account<>?2 OR cookie_hash<>?5))",
        params![scope.origin, scope.account, scope.workspace, scope.slug, scope.cookie_hash],
        |row| row.get(0),
    ).map_err(|error| error.to_string())?;
    if !valid {
        return Err("Verify this workspace before retaining a local draft".into());
    }
    Ok(())
}

fn retain(db: &Connection, scope: &BoundScope, key: &str, payload: &Value) -> Result<(), String> {
    validate_key(key)?;
    let source = key;
    let (key, payload) = super::identity::normalize(db, scope, key, payload)?;
    super::identity::ensure_writer(db, scope, source, &key)?;
    let encoded = serde_json::to_string(&payload).map_err(|error| error.to_string())?;
    if encoded.len() > MAX_DRAFT_BYTES {
        return Err("This local draft exceeds the 2 MiB limit".into());
    }
    known_scope(db, scope)?;
    if let Some(document_id) = key.strip_prefix("collaboration:") {
        use base64::Engine;
        if payload["version"] != 1
            || payload["kind"] != "collaboration"
            || !payload["dirty"].is_boolean()
            || !payload["canEdit"].is_boolean()
            || !payload["name"].as_str().is_some_and(|name| {
                name.starts_with(&format!("{document_id}:")) && name.len() < 160
            })
            || !payload["state"].as_str().is_some_and(|state| {
                base64::engine::general_purpose::STANDARD
                    .decode(state)
                    .is_ok()
            })
        {
            return Err("Invalid collaborative document draft".into());
        }
        if payload["dirty"] == false {
            ensure_scope_current(db, scope)?;
        }
    }
    // This only retains text supplied by the current browser session. It does not
    // renew a lease or grant permission to read cached data after access expires.
    db.execute(
        "INSERT INTO hosted_drafts VALUES(?,?,?,?,?,?) ON CONFLICT(origin,account,workspace,key) DO UPDATE SET payload=excluded.payload,saved=excluded.saved",
        params![scope.origin,scope.account,scope.workspace,key,encoded,now()],
    ).map_err(|_| "Could not retain your local draft. Check available disk space.".to_owned())?;
    super::identity::remember_writer(db, scope, source, &key)?;
    Ok(())
}

#[tauri::command]
pub async fn hosted_draft_write(
    window: WebviewWindow,
    request: WriteRequest,
) -> Result<(), String> {
    validate_key(&request.key)?;
    let state = authorize_window(&window)?;
    let scope = authorize_known(&window, &request.account_id, &request.workspace_id).await?;
    let mut db = state.db.lock().map_err(|error| error.to_string())?;
    let tx = db.transaction().map_err(|error| error.to_string())?;
    retain(&tx, &scope, &request.key, &request.payload)?;
    tx.commit().map_err(|error| error.to_string())
}

pub(super) async fn authorize_known(
    window: &WebviewWindow,
    account: &str,
    workspace: &str,
) -> Result<BoundScope, String> {
    let state = authorize_window(window)?;
    let origin = window
        .url()
        .map_err(|error| error.to_string())?
        .origin()
        .ascii_serialization();
    let token = browser_token(window).await?;
    let hash = fingerprint(&token);
    let slug: String = {
        let db = state.db.lock().map_err(|error| error.to_string())?;
        db.query_row("SELECT slug FROM hosted_known WHERE origin=? AND account=? AND workspace=? AND cookie_hash=?", params![origin,account,workspace,hash], |row| row.get(0))
            .map_err(|_| "Verify this workspace before retaining a local draft")?
    };
    if caller_origin(window, &slug)? != origin {
        return Err("Workspace changed".into());
    }
    let scope = BoundScope {
        origin,
        account: account.to_owned(),
        workspace: workspace.to_owned(),
        slug,
        cookie_hash: hash,
        token,
    };
    let db = state.db.lock().map_err(|error| error.to_string())?;
    known_scope(&db, &scope)?;
    Ok(scope)
}

#[tauri::command]
pub async fn hosted_draft_read(
    window: WebviewWindow,
    request: DraftRequest,
) -> Result<Option<Value>, String> {
    validate_key(&request.key)?;
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let state = authorize_window(&window)?;
    let db = state.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, &scope)?;
    let (key, _) = super::identity::normalize(&db, &scope, &request.key, &Value::Null)?;
    let encoded: Option<String> = db.query_row("SELECT payload FROM hosted_drafts WHERE origin=? AND account=? AND workspace=? AND key=?", params![scope.origin,scope.account,scope.workspace,key], |row| row.get(0)).optional().map_err(|error| error.to_string())?;
    encoded
        .map(|value| {
            serde_json::from_str(&value)
                .map_err(|_| "The retained local draft is invalid".to_owned())
        })
        .transpose()
}

#[tauri::command]
pub async fn hosted_draft_delete(
    window: WebviewWindow,
    request: DraftRequest,
) -> Result<(), String> {
    validate_key(&request.key)?;
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let state = authorize_window(&window)?;
    let db = state.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, &scope)?;
    let (key, _) = super::identity::normalize(&db, &scope, &request.key, &Value::Null)?;
    super::identity::ensure_writer(&db, &scope, &request.key, &key)?;
    db.execute(
        "DELETE FROM hosted_drafts WHERE origin=? AND account=? AND workspace=? AND key=?",
        params![scope.origin, scope.account, scope.workspace, key],
    )
    .map_err(|error| error.to_string())?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn comment_draft_keys_separate_replies_and_edits() {
        let story = "22222222-2222-4222-8222-222222222222";
        let comment = "33333333-3333-4333-8333-333333333333";
        for key in [
            format!("comment:{story}"),
            format!("comment:{story}:reply:{comment}"),
            format!("comment:{story}:edit:{comment}"),
        ] {
            assert!(validate_key(&key).is_ok());
        }
        for key in [
            format!("story:{story}:reply:{comment}"),
            format!("comment:{story}:delete:{comment}"),
            format!("comment:{story}:edit:invalid"),
        ] {
            assert!(validate_key(&key).is_err());
        }
    }

    #[test]
    fn expired_access_retains_only_supplied_draft_without_granting_reads() {
        let db = Connection::open_in_memory().unwrap();
        initialize(&db).unwrap();
        let scope = super::super::tests::scope();
        db.execute(
            "INSERT INTO hosted_known VALUES(?,?,?,?,?)",
            params![
                scope.origin,
                scope.account,
                scope.workspace,
                scope.slug,
                scope.cookie_hash
            ],
        )
        .unwrap();
        let key = "story:22222222-2222-4222-8222-222222222222";
        retain(&db, &scope, key, &serde_json::json!({"title":"draft"})).unwrap();
        assert!(ensure_scope_current(&db, &scope).is_err());
        let other = BoundScope {
            account: "another".into(),
            ..super::super::tests::scope()
        };
        assert!(retain(&db, &other, key, &Value::Null).is_err());
        assert!(retain(&db, &scope, "unknown:key", &Value::Null).is_err());
        assert_eq!(
            db.query_row("SELECT COUNT(*) FROM hosted_drafts", [], |row| row
                .get::<_, u64>(0))
                .unwrap(),
            1
        );
    }
}
