use super::*;

const MAX_CACHE_BYTES: usize = 16 * 1024 * 1024;

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct CacheRequest {
    account_id: String,
    workspace_id: String,
    version: u32,
}
#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct WriteRequest {
    account_id: String,
    workspace_id: String,
    version: u32,
    payload: Value,
}
#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ClearRequest {
    account_id: String,
}

fn validate(payload: &Value, scope: &BoundScope) -> Result<(), String> {
    if payload["version"] != 1
        || payload["accountId"].as_str() != Some(&scope.account)
        || payload["workspaceId"].as_str() != Some(&scope.workspace)
        || payload["workspaceSlug"].as_str() != Some(&scope.slug)
    {
        return Err("Downloaded data belongs to another account or workspace".into());
    }
    let queries = payload["queries"]
        .as_array()
        .ok_or("Invalid downloaded queries")?;
    if queries.len() > 2000
        || queries.iter().any(|entry| {
            !entry["queryKey"].is_array()
                || !entry["updatedAt"].is_u64()
                || entry.get("data").is_none()
        })
        || payload.to_string().len() > MAX_CACHE_BYTES
    {
        return Err("This workspace download exceeds the native cache limit".into());
    }
    Ok(())
}

#[tauri::command]
pub async fn hosted_cache_read(
    window: WebviewWindow,
    request: CacheRequest,
) -> Result<Option<Value>, String> {
    if request.version != 1 {
        return Err("Unsupported desktop cache version".into());
    }
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let state = authorize_window(&window)?;
    let db = state.db.lock().map_err(|error| error.to_string())?;
    binding(
        &db,
        &scope.origin,
        &scope.account,
        &scope.workspace,
        &scope.cookie_hash,
        now(),
    )?;
    let encoded:Option<String>=db.query_row("SELECT payload FROM hosted_cache WHERE origin=? AND account=? AND workspace=? AND version=1 AND saved>=?",params![scope.origin,scope.account,scope.workspace,now().saturating_sub(LEASE_MS)],|row|row.get(0)).optional().map_err(|error|error.to_string())?;
    encoded
        .map(|value| {
            serde_json::from_str(&value).map_err(|_| "The native query cache is invalid".into())
        })
        .transpose()
}

#[tauri::command]
pub async fn hosted_cache_write(
    window: WebviewWindow,
    mut request: WriteRequest,
) -> Result<(), String> {
    if request.version != 1 {
        return Err("Unsupported desktop cache version".into());
    }
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    validate(&request.payload, &scope)?;
    let saved = now();
    request.payload["savedAt"] = serde_json::json!(saved);
    let state = authorize_window(&window)?;
    let db = state.db.lock().map_err(|error| error.to_string())?;
    binding(
        &db,
        &scope.origin,
        &scope.account,
        &scope.workspace,
        &scope.cookie_hash,
        now(),
    )?;
    db.execute("INSERT INTO hosted_cache VALUES(?,?,?,1,?,?) ON CONFLICT(origin,account,workspace) DO UPDATE SET version=1,payload=excluded.payload,saved=excluded.saved",params![scope.origin,scope.account,scope.workspace,request.payload.to_string(),saved]).map_err(|_|"Could not save the native workspace cache. Check available disk space.".to_owned())?;
    Ok(())
}

#[tauri::command]
pub async fn hosted_cache_clear(
    window: WebviewWindow,
    request: ClearRequest,
) -> Result<(), String> {
    let state = authorize_window(&window)?;
    let origin = window
        .url()
        .map_err(|error| error.to_string())?
        .origin()
        .ascii_serialization();
    let mut db = state.db.lock().map_err(|error| error.to_string())?;
    let active: Option<String> = db
        .query_row(
            "SELECT account FROM hosted_active WHERE origin=?",
            [&origin],
            |row| row.get(0),
        )
        .optional()
        .map_err(|error| error.to_string())?;
    match active.as_deref() {
        None => {
            state.generation.fetch_add(1, Ordering::AcqRel);
            return Ok(());
        }
        Some(account) if account != request.account_id => return Err("Your account changed".into()),
        Some(_) => {}
    }
    state.generation.fetch_add(1, Ordering::AcqRel);
    let tx = db.transaction().map_err(|error| error.to_string())?;
    for table in [
        "hosted_cache",
        "hosted_bindings",
        "hosted_active",
        "hosted_notifications",
        "hosted_known",
        "hosted_results",
        "hosted_pages",
    ] {
        tx.execute(
            &format!("DELETE FROM {table} WHERE account=?"),
            [&request.account_id],
        )
        .map_err(|error| error.to_string())?;
    }
    tx.execute(
        "DELETE FROM hosted_files WHERE account=? AND uploaded=1",
        [&request.account_id],
    )
    .map_err(|error| error.to_string())?;
    tx.execute("DELETE FROM hosted_drafts WHERE account=? AND key LIKE 'collaboration:%' AND json_extract(payload,'$.dirty')=0",[&request.account_id]).map_err(|error|error.to_string())?;
    tx.commit().map_err(|error| error.to_string())?;
    window
        .set_badge_count(None)
        .map_err(|error| error.to_string())
}

#[cfg(test)]
pub(super) fn validate_for_test(payload: &Value, scope: &BoundScope) -> Result<(), String> {
    validate(payload, scope)
}
