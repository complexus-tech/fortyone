use super::{
    model::{now, Operation},
    runtime,
    store::Store,
    transport::decode,
    Runtime,
};
use rusqlite::{params, OptionalExtension};
use serde_json::{json, Value};
use tauri::{Emitter, WebviewWindow};

fn replace_id(value: &mut Value, old: &str, new: &str) {
    if let Some(map) = value.as_object_mut() {
        for key in ["storyId", "documentId", "parentId"] {
            if map.get(key).and_then(Value::as_str) == Some(old) {
                map.insert(key.into(), json!(new));
            }
        }
    }
}

fn resource_version(kind: &str, value: &Value) -> Option<String> {
    if kind.starts_with("document.") {
        value["revision"].as_u64().map(|v| v.to_string())
    } else {
        value["updatedAt"].as_str().map(str::to_owned)
    }
}

impl Store {
    pub fn acknowledge(
        &self,
        account: &str,
        workspace: &str,
        operation: &Operation,
        response: &Value,
    ) -> Result<(), String> {
        let attachment = operation.kind == "attachment.upload";
        let new_id = if attachment {
            &operation.resource_id
        } else {
            response["resourceId"]
                .as_str()
                .unwrap_or(&operation.resource_id)
        };
        uuid::Uuid::parse_str(new_id).map_err(|_| "Invalid resource response")?;
        let resource = response
            .get("resource")
            .filter(|v| v.is_object() && !attachment);
        let mut db = self.0.lock().map_err(|e| e.to_string())?;
        let tx = db.transaction().map_err(|e| e.to_string())?;
        let previous: Option<String> = tx
            .query_row(
                "SELECT data FROM snapshots WHERE account=? AND workspace=?",
                params![account, workspace],
                |r| r.get(0),
            )
            .optional()
            .map_err(|e| e.to_string())?;
        let mut snapshot = previous
            .map(|s| serde_json::from_str(&s))
            .transpose()
            .map_err(|e| e.to_string())?
            .unwrap_or_else(|| super::model::empty_workspace(workspace));
        if let Some(resource) = resource {
            let collection = if operation.kind.starts_with("story.") {
                "stories"
            } else if operation.kind.starts_with("document.") {
                "documents"
            } else {
                "comments"
            };
            if let Some(items) = snapshot[collection].as_array_mut() {
                items.retain(|item| {
                    item["id"].as_str() != Some(new_id)
                        && item["id"].as_str() != Some(&operation.resource_id)
                });
                items.push(resource.clone());
            }
        }
        if attachment {
            let mut file = response["resource"].clone();
            let id = response["resourceId"]
                .as_str()
                .ok_or("Missing uploaded attachment identity")?;
            uuid::Uuid::parse_str(id).map_err(|_| "Invalid uploaded attachment identity")?;
            if !file.is_object() {
                return Err("Missing uploaded attachment details".into());
            }
            file["id"] = json!(id);
            file["resourceId"] = json!(operation.resource_id);
            file["targetKind"] = operation.payload["targetKind"].clone();
            let mut attachments = snapshot["attachments"]
                .as_array()
                .cloned()
                .unwrap_or_default();
            attachments.retain(|file| file["id"].as_str() != Some(id));
            attachments.push(file);
            snapshot["attachments"] = json!(attachments);
        }
        tx.execute("INSERT INTO snapshots VALUES(?,?,?,?) ON CONFLICT(account,workspace) DO UPDATE SET data=excluded.data",params![account,workspace,snapshot.to_string(),now()]).map_err(|e|e.to_string())?;
        let pending: Vec<(String, String)> = {
            let mut query = tx
                .prepare(
                    "SELECT id,operation FROM operations WHERE account=? AND workspace=? AND id<>?",
                )
                .map_err(|e| e.to_string())?;
            let rows = query
                .query_map(params![account, workspace, operation.id], |r| {
                    Ok((r.get(0)?, r.get(1)?))
                })
                .map_err(|e| e.to_string())?;
            rows.collect::<Result<_, _>>().map_err(|e| e.to_string())?
        };
        for (id, encoded) in pending {
            let mut pending: Operation =
                serde_json::from_str(&encoded).map_err(|e| e.to_string())?;
            let same_resource = pending.resource_id == operation.resource_id;
            if same_resource {
                pending.resource_id = new_id.to_owned();
                if (operation.kind.ends_with(".create")
                    || pending.base_version == operation.base_version)
                    && !attachment
                {
                    if let Some(version) = response["operationVersion"].as_str().map(str::to_owned)
                    {
                        pending.base_version = Some(version);
                    }
                }
            }
            replace_id(&mut pending.payload, &operation.resource_id, new_id);
            tx.execute(
                "UPDATE operations SET operation=? WHERE id=?",
                params![
                    serde_json::to_string(&pending).map_err(|e| e.to_string())?,
                    id
                ],
            )
            .map_err(|e| e.to_string())?;
        }
        super::identity::record_creation(
            &tx,
            account,
            workspace,
            operation,
            new_id,
            response["operationVersion"].as_str(),
        )?;
        tx.execute(
            "DELETE FROM operations WHERE account=? AND workspace=? AND id=?",
            params![account, workspace, operation.id],
        )
        .map_err(|e| e.to_string())?;
        if attachment {
            tx.execute(
                "UPDATE attachments SET id=? WHERE account=? AND workspace=? AND id=?",
                params![
                    response["resourceId"]
                        .as_str()
                        .ok_or("Missing uploaded attachment identity")?,
                    account,
                    workspace,
                    operation.id
                ],
            )
            .map_err(|e| e.to_string())?;
        }
        tx.commit().map_err(|e| e.to_string())
    }
}

pub async fn synchronize(runtime: &Runtime, workspace_id: &str) -> Result<Value, String> {
    synchronize_for(runtime, workspace_id, None, false).await
}

async fn synchronize_for(
    runtime: &Runtime,
    workspace_id: &str,
    expected_account: Option<&str>,
    retry_credentials: bool,
) -> Result<Value, String> {
    let _guard = runtime.sync.lock().await;
    let account = match expected_account {
        Some(expected) => super::require_account(runtime, expected)?,
        None => runtime.store.account()?,
    };
    if retry_credentials {
        runtime.transport.retry_credentials(&account)?;
    }
    let token = runtime.transport.token(&account).await?;
    // Membership refresh precedes any queued write. Failure preserves the queue.
    let user = runtime.transport.get("/auth/me", &token).await?;
    if user["id"].as_str() != Some(&account) {
        return Err("Your account changed. Sign in again before synchronizing.".into());
    }
    let workspaces = runtime.transport.get("/workspaces", &token).await?;
    runtime.store.save_account(&user, &workspaces)?;
    let workspace = runtime.store.membership(&account, workspace_id)?;
    let slug = workspace["slug"]
        .as_str()
        .filter(|s| !s.is_empty() && s.bytes().all(|b| b.is_ascii_alphanumeric() || b == b'-'))
        .ok_or("Invalid workspace address")?;
    let path = format!("/workspaces/{slug}/desktop");
    // Verify current per-workspace access (including SSO) before replaying writes.
    super::download::fetch(
        runtime,
        &account,
        workspace_id,
        &[("storyLimit", "50".into()), ("documentLimit", "1".into())],
    )
    .await?;
    // Re-read after each acknowledgement: generated IDs and dependent versions
    // are rewritten atomically together with the acknowledged local snapshot.
    loop {
        let operations = runtime.store.operations(&account, workspace_id)?;
        let Some(mut operation) = super::model::next_operation(operations) else {
            break;
        };
        let mut payload = operation.payload.clone();
        if operation.kind == "attachment.upload" {
            use base64::Engine;
            let bytes: Vec<u8> = runtime
                .store
                .0
                .lock()
                .map_err(|e| e.to_string())?
                .query_row(
                    "SELECT bytes FROM attachments WHERE id=? AND account=? AND workspace=?",
                    params![operation.id, account, workspace_id],
                    |r| r.get(0),
                )
                .map_err(|_| "The staged attachment could not be read".to_owned())?;
            payload["data"] = json!(base64::engine::general_purpose::STANDARD.encode(bytes));
            payload["contentType"] = json!("application/octet-stream");
            if let Some(object) = payload.as_object_mut() {
                object.remove("attachmentId");
                object.remove("size");
            }
        }
        operation.attempted = true;
        runtime.store.mark_operation(&account, &operation)?;
        let response=runtime.transport.request(reqwest::Method::POST,&format!("{path}/operations"),&token)?.json(&json!({"id":operation.id,"kind":operation.kind,"resourceId":operation.resource_id,"payload":payload,"baseVersion":operation.base_version})).send().await;
        let response = match response {
            Ok(value) => value,
            Err(_) => {
                operation.error=Some("Connection interrupted. This change will be reconciled using the same operation ID.".into());
                runtime.store.mark_operation(&account, &operation)?;
                return Err("You are offline. Changes are saved on this device.".into());
            }
        };
        let status = response.status().as_u16();
        match decode(response).await {
            Ok(result) => runtime
                .store
                .acknowledge(&account, workspace_id, &operation, &result)?,
            Err(error) => {
                operation.status = match status {
                    409 | 412 => "conflict",
                    400 | 401 | 403 | 404 | 413 | 422 => "blocked",
                    _ => "pending",
                }
                .into();
                operation.error = Some(error.clone());
                runtime.store.mark_operation(&account, &operation)?;
                if operation.status == "pending" {
                    return Err(error);
                }
                continue;
            }
        }
    }
    let previous = runtime.store.server_snapshot(&account, workspace_id)?;
    // Revalidate every downloaded identity, including items outside the recent
    // page. Explicit removals prevent deleted or revoked data lingering offline.
    for (collection, parameter, size) in [
        ("stories", "storyIds", 50usize),
        ("documents", "documentIds", 1usize),
    ] {
        let ids: Vec<String> = previous[collection]
            .as_array()
            .into_iter()
            .flatten()
            .filter_map(|v| v["id"].as_str().map(str::to_owned))
            .collect();
        for chunk in ids.chunks(size) {
            super::download::fetch(
                runtime,
                &account,
                workspace_id,
                &[("scope", "selected".into()), (parameter, chunk.join(","))],
            )
            .await?;
        }
    }
    let snapshot = runtime.store.server_snapshot(&account, workspace_id)?;
    for mut op in runtime.store.operations(&account, workspace_id)? {
        if op.status != "conflict" {
            continue;
        }
        let collection = if op.kind.starts_with("story.") {
            "stories"
        } else {
            "documents"
        };
        if let Some(resource) = snapshot[collection].as_array().and_then(|values| {
            values
                .iter()
                .find(|v| v["id"].as_str() == Some(&op.resource_id))
        }) {
            op.server_version = resource_version(&op.kind, resource);
            op.server_value = Some(resource.clone());
            runtime.store.mark_operation(&account, &op)?;
        }
    }
    runtime.store.workspace(&account, workspace_id)
}

#[tauri::command]
pub async fn desktop_sync(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    retry_credentials: Option<bool>,
) -> Result<Value, String> {
    let runtime = runtime(&window)?;
    let result = synchronize_for(
        &runtime,
        &workspace_id,
        Some(&expected_account_id),
        retry_credentials == Some(true),
    )
    .await;
    if result.is_ok() {
        let _ = window.emit(
            "desktop-refresh-request",
            serde_json::json!({"workspaceId":workspace_id}),
        );
    }
    let _ = window.emit("desktop-state-changed", ());
    result
}
