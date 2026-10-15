//! Durable identities bridge provisional offline creations and server IDs.
//! Acknowledgements and draft consumption share the outbox transaction.
use super::{
    model::{now, validate_operation, Operation},
    store::Store,
};
use rusqlite::{params, Connection, OptionalExtension};
use serde::Deserialize;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};

#[cfg(test)]
mod tests;

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ConsumedDraft {
    pub resource_id: String,
    pub kind: String,
}

struct Alias {
    resource: String,
    kind: String,
    version: Option<String>,
    creation_hash: String,
}

fn alias(
    db: &Connection,
    account: &str,
    workspace: &str,
    resource: &str,
) -> Result<Option<Alias>, String> {
    db.query_row(
        "SELECT resource,kind,version,creation_hash FROM identity_aliases WHERE account=? AND workspace=? AND provisional=?",
        params![account,workspace,resource],
        |row| Ok(Alias { resource: row.get(0)?, kind: row.get(1)?, version: row.get(2)?, creation_hash: row.get(3)? }),
    ).optional().map_err(|error| error.to_string())
}

pub fn resolve_resource(
    db: &Connection,
    account: &str,
    workspace: &str,
    resource: &str,
) -> Result<String, String> {
    let mut current = resource.to_owned();
    for _ in 0..8 {
        match alias(db, account, workspace, &current)? {
            Some(mapped) if mapped.resource != current => current = mapped.resource,
            _ => return Ok(current),
        }
    }
    Err("This saved item has an invalid identity mapping".into())
}

fn normalize_references(
    db: &Connection,
    account: &str,
    workspace: &str,
    payload: &mut Value,
) -> Result<(), String> {
    if let Some(object) = payload.as_object_mut() {
        for field in ["resourceId", "storyId", "documentId", "parentId"] {
            if let Some(id) = object.get(field).and_then(Value::as_str) {
                let resolved = resolve_resource(db, account, workspace, id)?;
                object.insert(field.into(), json!(resolved));
            }
        }
    }
    Ok(())
}

fn creation_version(
    db: &Connection,
    account: &str,
    workspace: &str,
    marker: &str,
    kind: &str,
) -> Result<Option<String>, String> {
    let Some(provisional) = marker.strip_prefix("pending-create:") else {
        return Ok(None);
    };
    let Some(mapped) = alias(db, account, workspace, provisional)? else {
        return Ok(None);
    };
    if mapped.kind != kind {
        return Err("The saved creation belongs to another resource type".into());
    }
    mapped
        .version
        .map(Some)
        .ok_or_else(|| "Reconnect to confirm the created item's original version".into())
}

pub fn normalize_draft(
    db: &Connection,
    account: &str,
    workspace: &str,
    resource: &str,
    kind: &str,
    payload: &Value,
) -> Result<(String, Value), String> {
    let resolved = resolve_resource(db, account, workspace, resource)?;
    let mut payload = payload.clone();
    normalize_references(db, account, workspace, &mut payload)?;
    if let Some(marker) = payload.get("baseVersion").and_then(Value::as_str) {
        if let Some(version) = creation_version(db, account, workspace, marker, kind)? {
            payload["baseVersion"] = json!(version);
        }
    }
    Ok((resolved, payload))
}

fn payload_hash(payload: &Value) -> String {
    format!("{:x}", Sha256::digest(payload.to_string().as_bytes()))
}

fn creation_alias(
    db: &Connection,
    account: &str,
    workspace: &str,
    resource: &str,
) -> Result<Option<Alias>, String> {
    // A recovered creation draft may already have had its payload ID remapped.
    // Both names identify the same acknowledged creation and must deduplicate.
    db.query_row(
        "SELECT resource,kind,version,creation_hash FROM identity_aliases WHERE account=? AND workspace=? AND (provisional=? OR resource=?) LIMIT 1",
        params![account,workspace,resource,resource],
        |row| Ok(Alias { resource: row.get(0)?, kind: row.get(1)?, version: row.get(2)?, creation_hash: row.get(3)? }),
    ).optional().map_err(|error| error.to_string())
}

pub(super) fn record_creation(
    db: &Connection,
    account: &str,
    workspace: &str,
    operation: &Operation,
    resource: &str,
    version: Option<&str>,
) -> Result<(), String> {
    let Some(kind) = operation.kind.strip_suffix(".create") else {
        return Ok(());
    };
    let hash = payload_hash(&operation.payload);
    db.execute(
        "INSERT INTO identity_aliases(account,workspace,provisional,resource,kind,version,creation_hash) VALUES(?,?,?,?,?,?,?) ON CONFLICT(account,workspace,provisional) DO NOTHING",
        params![account,workspace,operation.resource_id,resource,kind,version,hash],
    ).map_err(|error| error.to_string())?;
    let stored = alias(db, account, workspace, &operation.resource_id)?
        .ok_or("Missing creation identity")?;
    if stored.resource != resource
        || stored.kind != kind
        || stored.version.as_deref() != version
        || stored.creation_hash != hash
    {
        return Err("This creation was already acknowledged with a different identity".into());
    }
    let drafts: Vec<(String, String, String)> = {
        let mut query = db
            .prepare("SELECT resource,kind,payload FROM drafts WHERE account=? AND workspace=?")
            .map_err(|error| error.to_string())?;
        let rows = query
            .query_map(params![account, workspace], |row| {
                Ok((row.get(0)?, row.get(1)?, row.get(2)?))
            })
            .map_err(|error| error.to_string())?;
        rows.collect::<Result<_, _>>()
            .map_err(|error| error.to_string())?
    };
    for (old_resource, draft_kind, encoded) in drafts {
        let payload: Value = serde_json::from_str(&encoded).map_err(|error| error.to_string())?;
        let (new_resource, normalized) =
            normalize_draft(db, account, workspace, &old_resource, &draft_kind, &payload)?;
        if old_resource != new_resource || normalized != payload {
            db.execute("UPDATE drafts SET resource=?,payload=? WHERE account=? AND workspace=? AND resource=? AND kind=?",params![new_resource,normalized.to_string(),account,workspace,old_resource,draft_kind]).map_err(|error| error.to_string())?;
        }
    }
    Ok(())
}

fn consume_draft(
    db: &Connection,
    account: &str,
    workspace: &str,
    draft: Option<&ConsumedDraft>,
) -> Result<(), String> {
    let Some(draft) = draft else {
        return Ok(());
    };
    if !matches!(draft.kind.as_str(), "story" | "document" | "comment")
        || draft.resource_id.is_empty()
        || draft.resource_id.len() > 128
    {
        return Err("Invalid submitted draft".into());
    }
    let resolved = resolve_resource(db, account, workspace, &draft.resource_id)?;
    db.execute(
        "DELETE FROM drafts WHERE account=? AND workspace=? AND kind=? AND resource IN (?,?)",
        params![account, workspace, draft.kind, draft.resource_id, resolved],
    )
    .map_err(|error| error.to_string())?;
    Ok(())
}

fn verify_workspace(db: &Connection, account: &str, workspace: &str) -> Result<(), String> {
    super::require_account_in(db, account)?;
    let earliest = now().saturating_sub(7 * 24 * 3600);
    let valid: bool = db.query_row(
        "SELECT EXISTS(SELECT 1 FROM accounts WHERE id=?1 AND verified>=?3 AND EXISTS(SELECT 1 FROM json_each(accounts.workspaces) WHERE json_extract(value,'$.id')=?2)) AND NOT EXISTS(SELECT 1 FROM workspace_access WHERE account=?1 AND workspace=?2 AND (denied IS NOT NULL OR verified<?3))",
        params![account,workspace,earliest], |row| row.get(0),
    ).map_err(|error| error.to_string())?;
    if !valid {
        return Err("Account or workspace access changed. Your draft is retained.".into());
    }
    Ok(())
}

impl Store {
    pub fn enqueue_consuming(
        &self,
        account: &str,
        workspace: &str,
        operation: &Operation,
        consumed: Option<&ConsumedDraft>,
    ) -> Result<(), String> {
        validate_operation(operation)?;
        let mut db = self.0.lock().map_err(|error| error.to_string())?;
        let tx = db.transaction().map_err(|error| error.to_string())?;
        verify_workspace(&tx, account, workspace)?;
        let mut normalized = operation.clone();
        normalize_references(&tx, account, workspace, &mut normalized.payload)?;
        if let Some(kind) = normalized.kind.strip_suffix(".create") {
            let mut duplicate = false;
            if let Some(mapped) = creation_alias(&tx, account, workspace, &normalized.resource_id)?
            {
                if mapped.kind != kind || mapped.creation_hash != payload_hash(&normalized.payload)
                {
                    return Err("This item was already created. Open the saved item to apply further edits; your draft is retained.".into());
                }
                duplicate = true;
            } else {
                let existing: Option<String> = tx.query_row(
                    "SELECT operation FROM operations WHERE account=? AND workspace=? AND json_extract(operation,'$.kind')=? AND json_extract(operation,'$.resourceId')=? LIMIT 1",
                    params![account,workspace,normalized.kind,normalized.resource_id], |row| row.get(0),
                ).optional().map_err(|error| error.to_string())?;
                if let Some(encoded) = existing {
                    let existing: Operation =
                        serde_json::from_str(&encoded).map_err(|error| error.to_string())?;
                    if existing.payload != normalized.payload {
                        return Err("This item is already queued. Open it to edit the pending creation; your draft is retained.".into());
                    }
                    duplicate = true;
                }
            }
            if duplicate {
                consume_draft(&tx, account, workspace, consumed)?;
                return tx.commit().map_err(|error| error.to_string());
            }
        } else {
            normalized.resource_id =
                resolve_resource(&tx, account, workspace, &operation.resource_id)?;
        }
        if let Some(marker) = operation
            .base_version
            .as_deref()
            .filter(|value| value.starts_with("pending-create:"))
        {
            let expected = format!("pending-create:{}", operation.resource_id);
            let kind = operation
                .kind
                .strip_suffix(".update")
                .ok_or("Invalid pending creation version")?;
            if marker != expected {
                return Err("The original offline creation could not be found".into());
            }
            if let Some(version) = creation_version(&tx, account, workspace, marker, kind)? {
                normalized.base_version = Some(version);
            } else {
                let exists: bool = tx.query_row("SELECT EXISTS(SELECT 1 FROM operations WHERE account=? AND workspace=? AND json_extract(operation,'$.kind')=? AND json_extract(operation,'$.resourceId')=?)",params![account,workspace,format!("{kind}.create"),operation.resource_id],|row|row.get(0)).map_err(|error| error.to_string())?;
                if !exists {
                    return Err("The original offline creation could not be found".into());
                }
            }
        }
        tx.execute(
            "INSERT INTO operations VALUES(?,?,?,?)",
            params![
                normalized.id,
                account,
                workspace,
                serde_json::to_string(&normalized).map_err(|error| error.to_string())?
            ],
        )
        .map_err(|error| error.to_string())?;
        consume_draft(&tx, account, workspace, consumed)?;
        tx.commit().map_err(|error| error.to_string())
    }
}
