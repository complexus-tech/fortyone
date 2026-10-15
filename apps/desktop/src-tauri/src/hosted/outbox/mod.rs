//! Durable mutations for the existing Projects editors. Credentials remain in
//! the native caller scope; only operation results cross the bridge.
mod replay;
pub(super) mod store;
#[cfg(test)]
mod tests;

use super::*;
use crate::operations::{next_operation, validate_operation, Operation};
use serde_json::json;

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ScopeRequest {
    account_id: String,
    workspace_id: String,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct EnqueueRequest {
    pub(super) account_id: String,
    pub(super) workspace_id: String,
    pub(super) id: String,
    pub(super) kind: String,
    pub(super) resource_id: String,
    pub(super) payload: Value,
    pub(super) base_version: Option<String>,
    pub(super) consumed_draft_key: Option<String>,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ResolveRequest {
    account_id: String,
    workspace_id: String,
    operation_id: String,
    resolution: String,
    payload: Option<Value>,
    base_version: Option<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct Submission {
    pub(super) operation_id: String,
    pub(super) resource_id: String,
    pub(super) pending: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    resource: Option<Value>,
}

#[derive(Serialize)]
pub struct Outbox {
    pub(super) operations: Vec<Operation>,
    pub(super) completed: Vec<Value>,
}

pub(super) fn initialize(db: &Connection) -> Result<(), String> {
    db.execute_batch("CREATE TABLE IF NOT EXISTS hosted_inputs(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,id TEXT NOT NULL,operation_id TEXT NOT NULL,input_hash TEXT NOT NULL,resource_id TEXT NOT NULL,PRIMARY KEY(origin,account,workspace,id)); CREATE TABLE IF NOT EXISTS hosted_discarded(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,id TEXT NOT NULL,PRIMARY KEY(origin,account,workspace,id));").map_err(|error|error.to_string())
}

#[tauri::command]
pub async fn hosted_enqueue_operation(
    window: WebviewWindow,
    request: EnqueueRequest,
) -> Result<Submission, String> {
    let state = authorize_window(&window)?;
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let mut db = state.db.lock().map_err(|error| error.to_string())?;
    let tx = db.transaction().map_err(|error| error.to_string())?;
    ensure_scope_current(&tx, &scope)?;
    let result = store::enqueue(&tx, &scope, &request)?;
    tx.commit().map_err(|error| error.to_string())?;
    Ok(result)
}

#[tauri::command]
pub async fn hosted_read_outbox(
    window: WebviewWindow,
    request: ScopeRequest,
) -> Result<Outbox, String> {
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let state = authorize_window(&window)?;
    let db = state.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, &scope)?;
    store::read(&db, &scope)
}

#[tauri::command]
pub async fn hosted_sync_outbox(
    window: WebviewWindow,
    request: ScopeRequest,
) -> Result<Outbox, String> {
    let state = authorize_window(&window)?;
    let _guard = state.outbox.lock().await;
    // Resolve the caller only after acquiring the replay lock: a queued request
    // may have outlived the account or workspace that submitted it.
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    replay::synchronize(&window, &state, &scope).await?;
    let db = state.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, &scope)?;
    store::read(&db, &scope)
}

#[tauri::command]
pub async fn hosted_resolve_operation(
    window: WebviewWindow,
    request: ResolveRequest,
) -> Result<Outbox, String> {
    let state = authorize_window(&window)?;
    let _guard = state.outbox.lock().await;
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let mut db = state.db.lock().map_err(|error| error.to_string())?;
    let tx = db.transaction().map_err(|error| error.to_string())?;
    ensure_scope_current(&tx, &scope)?;
    store::resolve(&tx, &scope, &request)?;
    tx.commit().map_err(|error| error.to_string())?;
    store::read(&db, &scope)
}
