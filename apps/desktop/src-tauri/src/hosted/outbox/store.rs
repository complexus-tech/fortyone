use super::*;

fn digest(value: &Value) -> String {
    fingerprint(&value.to_string())
}
fn command_hash(operation: &Operation) -> String {
    digest(
        &json!({"kind":operation.kind,"resourceId":operation.resource_id,"payload":operation.payload,"baseVersion":operation.base_version}),
    )
}
fn creation_hash(kind: &str, payload: &Value) -> String {
    digest(&json!({"kind":kind,"payload":payload}))
}

fn read_one(db: &Connection, scope: &BoundScope, id: &str) -> Result<Option<Operation>, String> {
    let encoded:Option<String>=db.query_row("SELECT operation FROM hosted_operations WHERE origin=? AND account=? AND workspace=? AND id=?",params![scope.origin,scope.account,scope.workspace,id],|row|row.get(0)).optional().map_err(|error|error.to_string())?;
    encoded
        .map(|value| serde_json::from_str(&value).map_err(|error| error.to_string()))
        .transpose()
}

pub(in crate::hosted) fn read(db: &Connection, scope: &BoundScope) -> Result<Outbox, String> {
    let mut statement=db.prepare("SELECT operation FROM hosted_operations WHERE origin=? AND account=? AND workspace=? ORDER BY rowid").map_err(|error|error.to_string())?;
    let encoded = statement
        .query_map(
            params![scope.origin, scope.account, scope.workspace],
            |row| row.get::<_, String>(0),
        )
        .map_err(|error| error.to_string())?;
    let operations = encoded
        .map(|value| {
            serde_json::from_str(&value.map_err(|error| error.to_string())?)
                .map_err(|error| error.to_string())
        })
        .collect::<Result<Vec<_>, String>>()?;
    let mut statement=db.prepare("SELECT result FROM hosted_results WHERE origin=? AND account=? AND workspace=? ORDER BY rowid").map_err(|error|error.to_string())?;
    let encoded = statement
        .query_map(
            params![scope.origin, scope.account, scope.workspace],
            |row| row.get::<_, String>(0),
        )
        .map_err(|error| error.to_string())?;
    let completed = encoded
        .map(|value| {
            serde_json::from_str(&value.map_err(|error| error.to_string())?)
                .map_err(|error| error.to_string())
        })
        .collect::<Result<Vec<_>, String>>()?;
    Ok(Outbox {
        operations,
        completed,
    })
}

fn alias(
    db: &Connection,
    scope: &BoundScope,
    id: &str,
) -> Result<Option<(String, Option<String>, String, String)>, String> {
    db.query_row("SELECT resource,version,kind,creation_hash FROM hosted_aliases WHERE origin=? AND account=? AND workspace=? AND provisional=?",params![scope.origin,scope.account,scope.workspace,id],|row|Ok((row.get(0)?,row.get(1)?,row.get(2)?,row.get(3)?))).optional().map_err(|error|error.to_string())
}

pub(in crate::hosted) fn resolve_resource(
    db: &Connection,
    scope: &BoundScope,
    id: &str,
) -> Result<String, String> {
    Ok(alias(db, scope, id)?
        .map(|(resource, _, _, _)| resource)
        .unwrap_or_else(|| id.to_owned()))
}

fn consume(db: &Connection, scope: &BoundScope, key: Option<&str>) -> Result<(), String> {
    if let Some(key) = key {
        super::super::drafts::validate_key(key)?;
        db.execute(
            "DELETE FROM hosted_drafts WHERE origin=? AND account=? AND workspace=? AND key=?",
            params![scope.origin, scope.account, scope.workspace, key],
        )
        .map_err(|error| error.to_string())?;
    }
    Ok(())
}

fn remember_input(
    db: &Connection,
    scope: &BoundScope,
    id: &str,
    operation_id: &str,
    hash: &str,
    resource: &str,
) -> Result<(), String> {
    db.execute(
        "INSERT INTO hosted_inputs VALUES(?,?,?,?,?,?,?)",
        params![
            scope.origin,
            scope.account,
            scope.workspace,
            id,
            operation_id,
            hash,
            resource
        ],
    )
    .map_err(|error| error.to_string())?;
    Ok(())
}

pub(in crate::hosted) fn enqueue(
    db: &Connection,
    scope: &BoundScope,
    request: &EnqueueRequest,
) -> Result<Submission, String> {
    uuid::Uuid::parse_str(&request.id).map_err(|_| "Invalid operation identifier")?;
    if !matches!(
        request.kind.as_str(),
        "story.create" | "story.update" | "comment.create" | "attachment.upload"
    ) {
        return Err("This editor change is not supported offline".into());
    }
    let mut operation = Operation {
        id: request.id.clone(),
        kind: request.kind.clone(),
        resource_id: request.resource_id.clone(),
        payload: request.payload.clone(),
        base_version: request.base_version.clone(),
        status: "pending".into(),
        error: None,
        server_version: None,
        server_value: None,
        attempted: false,
    };
    validate_operation(&operation)?;
    if operation.kind == "attachment.upload" {
        let id = operation.payload["attachmentId"]
            .as_str()
            .ok_or("Stage the attachment before queuing it")?;
        let retained:bool=db.query_row("SELECT EXISTS(SELECT 1 FROM hosted_files WHERE origin=? AND account=? AND workspace=? AND id=? AND hash=? AND name=? AND mime=?)",params![scope.origin,scope.account,scope.workspace,id,operation.payload["contentHash"].as_str(),operation.payload["fileName"].as_str(),operation.payload["contentType"].as_str()],|row|row.get(0)).map_err(|error|error.to_string())?;
        if !retained || request.id != id || operation.payload["targetKind"] != "story" {
            return Err("Stage the attachment before queuing it".into());
        }
    }
    let hash = command_hash(&operation);
    let previous:Option<(String,String,String)>=db.query_row("SELECT operation_id,input_hash,resource_id FROM hosted_inputs WHERE origin=? AND account=? AND workspace=? AND id=?",params![scope.origin,scope.account,scope.workspace,request.id],|row|Ok((row.get(0)?,row.get(1)?,row.get(2)?))).optional().map_err(|error|error.to_string())?;
    if let Some((id, previous_hash, resource)) = previous {
        if previous_hash != hash {
            return Err("This operation ID was already used for a different change".into());
        }
        let discarded:bool=db.query_row("SELECT EXISTS(SELECT 1 FROM hosted_discarded WHERE origin=? AND account=? AND workspace=? AND id=?)",params![scope.origin,scope.account,scope.workspace,id],|row|row.get(0)).map_err(|error|error.to_string())?;
        if discarded {
            return Err(
                "This change was discarded or replaced. Reopen the item before saving again."
                    .into(),
            );
        }
        let pending = read_one(db, scope, &id)?.is_some();
        return Ok(Submission {
            operation_id: id,
            resource_id: resource,
            pending,
            resource: None,
        });
    }
    if operation.kind.ends_with(".create") {
        if let Some((resource, _, kind, previous_hash)) = alias(db, scope, &operation.resource_id)?
        {
            if kind != operation.kind
                || previous_hash != creation_hash(&operation.kind, &operation.payload)
            {
                return Err("This draft was already created with different content".into());
            }
            remember_input(db, scope, &operation.id, &operation.id, &hash, &resource)?;
            return Ok(Submission {
                operation_id: operation.id,
                resource_id: resource,
                pending: false,
                resource: None,
            });
        }
        if let Some(existing) = read(db, scope)?.operations.into_iter().find(|value| {
            value.kind.ends_with(".create") && value.resource_id == operation.resource_id
        }) {
            if existing.kind != operation.kind || existing.payload != operation.payload {
                return Err("This draft already has a different pending creation".into());
            }
            remember_input(
                db,
                scope,
                &operation.id,
                &existing.id,
                &hash,
                &existing.resource_id,
            )?;
            return Ok(Submission {
                operation_id: existing.id,
                resource_id: existing.resource_id,
                pending: true,
                resource: None,
            });
        }
    } else if let Some((resource, version, _, _)) = alias(db, scope, &operation.resource_id)? {
        if operation.base_version.as_deref()
            == Some(&format!("pending-create:{}", operation.resource_id))
        {
            operation.base_version = version;
        }
        operation.resource_id = resource;
    }
    for key in ["storyId", "parentId"] {
        if let Some(id) = operation.payload[key].as_str().map(str::to_owned) {
            if let Some((resource, _, _, _)) = alias(db, scope, &id)? {
                operation.payload[key] = json!(resource);
            }
        }
    }
    if let Some(base) = operation
        .base_version
        .as_deref()
        .and_then(|base| base.strip_prefix("pending-create:"))
    {
        if base != operation.resource_id
            || !read(db, scope)?.operations.iter().any(|other| {
                other.kind == "story.create" && other.resource_id == operation.resource_id
            })
        {
            return Err("This local story creation is no longer pending".into());
        }
    }
    validate_operation(&operation)?;
    remember_input(
        db,
        scope,
        &operation.id,
        &operation.id,
        &hash,
        &operation.resource_id,
    )?;
    save(db, scope, &operation)?;
    consume(db, scope, request.consumed_draft_key.as_deref())?;
    Ok(Submission {
        operation_id: operation.id,
        resource_id: operation.resource_id,
        pending: true,
        resource: None,
    })
}

pub(super) fn save(
    db: &Connection,
    scope: &BoundScope,
    operation: &Operation,
) -> Result<(), String> {
    db.execute("INSERT INTO hosted_operations VALUES(?,?,?,?,?) ON CONFLICT(origin,account,workspace,id) DO UPDATE SET operation=excluded.operation",params![scope.origin,scope.account,scope.workspace,operation.id,serde_json::to_string(operation).map_err(|error|error.to_string())?]).map_err(|_|"Could not save this change locally. Check available disk space.".to_owned())?;
    Ok(())
}

pub(in crate::hosted) fn acknowledge(
    db: &Connection,
    scope: &BoundScope,
    operation: &Operation,
    result: &Value,
) -> Result<(), String> {
    let resource = result["resourceId"]
        .as_str()
        .ok_or("Missing saved resource identity")?;
    uuid::Uuid::parse_str(resource).map_err(|_| "Invalid saved resource identity")?;
    let version = result["operationVersion"].as_str();
    if operation.kind == "story.create" || operation.kind == "story.update" {
        if version.is_none_or(str::is_empty) {
            return Err(
                "Missing immutable saved story version; retry to reconcile this operation".into(),
            );
        }
    }
    let creation = operation.kind.ends_with(".create");
    let attachment = operation.kind == "attachment.upload";
    let provisional = if attachment {
        operation.payload["attachmentId"]
            .as_str()
            .ok_or("Missing retained attachment identity")?
    } else {
        &operation.resource_id
    };
    for mut pending in read(db, scope)?.operations {
        if attachment || pending.id == operation.id || pending.attempted {
            continue;
        }
        if pending.resource_id == operation.resource_id {
            pending.resource_id = resource.into();
            if creation || pending.base_version == operation.base_version {
                if let Some(version) = version {
                    pending.base_version = Some(version.into());
                }
            }
        }
        for key in ["storyId", "parentId"] {
            if pending.payload[key].as_str() == Some(&operation.resource_id) {
                pending.payload[key] = json!(resource);
            }
        }
        save(db, scope, &pending)?;
    }
    if creation {
        db.execute("INSERT INTO hosted_aliases VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(origin,account,workspace,provisional) DO UPDATE SET resource=excluded.resource,version=excluded.version",params![scope.origin,scope.account,scope.workspace,operation.resource_id,resource,operation.kind,version,creation_hash(&operation.kind,&operation.payload)]).map_err(|error|error.to_string())?;
        super::super::identity::migrate(db, scope)?;
    }
    if operation.kind == "story.create" {
        db.execute("UPDATE hosted_files SET story=? WHERE origin=? AND account=? AND workspace=? AND story=?",params![resource,scope.origin,scope.account,scope.workspace,operation.resource_id]).map_err(|error|error.to_string())?;
    }
    if attachment {
        let changed=db.execute("UPDATE hosted_files SET server_id=?,uploaded=1 WHERE origin=? AND account=? AND workspace=? AND id=? AND story=?",params![resource,scope.origin,scope.account,scope.workspace,provisional,operation.resource_id]).map_err(|error|error.to_string())?;
        if changed != 1 {
            return Err(
                "The retained upload is missing; its acknowledgement cannot be saved".into(),
            );
        }
    }
    // Keep bounded refreshed data for immediate React Query reconciliation. The
    // durable identity/version receipt remains useful when a large body is omitted.
    let refreshed = result
        .get("resource")
        .filter(|value| !value.is_null())
        .filter(|value| value.to_string().len() <= 64 * 1024);
    let mut acknowledgement = json!({"operationId":operation.id,"kind":operation.kind,"provisionalId":provisional,"resourceId":resource,"operationVersion":version,"resource":refreshed});
    if attachment {
        acknowledgement["storyId"] = json!(operation.resource_id);
        if refreshed.is_none() {
            acknowledgement["resource"] = json!({"id":resource,"filename":operation.payload["fileName"],"mimeType":operation.payload["contentType"],"size":operation.payload["size"],"createdAt":operation.payload["createdAt"],"uploadedBy":scope.account,"url":""});
        }
    }
    db.execute(
        "INSERT OR REPLACE INTO hosted_results VALUES(?,?,?,?,?,?)",
        params![
            scope.origin,
            scope.account,
            scope.workspace,
            operation.id,
            acknowledgement.to_string(),
            now()
        ],
    )
    .map_err(|error| error.to_string())?;
    if attachment {
        db.execute("UPDATE hosted_inputs SET resource_id=? WHERE origin=? AND account=? AND workspace=? AND operation_id=?",params![resource,scope.origin,scope.account,scope.workspace,operation.id]).map_err(|error|error.to_string())?;
    } else {
        db.execute("UPDATE hosted_inputs SET resource_id=? WHERE origin=? AND account=? AND workspace=? AND resource_id=?",params![resource,scope.origin,scope.account,scope.workspace,operation.resource_id]).map_err(|error|error.to_string())?;
    }
    db.execute("DELETE FROM hosted_results WHERE rowid NOT IN (SELECT rowid FROM hosted_results ORDER BY rowid DESC LIMIT 200)",[]).map_err(|error|error.to_string())?;
    db.execute(
        "DELETE FROM hosted_operations WHERE origin=? AND account=? AND workspace=? AND id=?",
        params![scope.origin, scope.account, scope.workspace, operation.id],
    )
    .map_err(|error| error.to_string())?;
    // Query snapshots are refreshed by the existing React Query invalidation;
    // stale authoritative entries must not resurrect after a native ACK.
    db.execute(
        "DELETE FROM hosted_cache WHERE origin=? AND account=? AND workspace=?",
        params![scope.origin, scope.account, scope.workspace],
    )
    .map_err(|error| error.to_string())?;
    Ok(())
}

pub(super) fn resolve(
    db: &Connection,
    scope: &BoundScope,
    request: &ResolveRequest,
) -> Result<(), String> {
    let mut operation = read_one(db, scope, &request.operation_id)?
        .ok_or("This pending change no longer exists")?;
    match request.resolution.as_str() {
        "discard" => {
            if operation.attempted && operation.status == "pending" {
                return Err("This change may already be saved. Reconnect to reconcile it before discarding.".into());
            }
            if operation.kind.ends_with(".create")
                && read(db, scope)?.operations.iter().any(|other| {
                    other.id != operation.id
                        && (other.resource_id == operation.resource_id
                            || ["storyId", "parentId"].iter().any(|key| {
                                other.payload[*key].as_str() == Some(&operation.resource_id)
                            }))
                })
            {
                return Err(
                    "Resolve this item's dependent changes before discarding its creation".into(),
                );
            }
            db.execute("DELETE FROM hosted_operations WHERE origin=? AND account=? AND workspace=? AND id=?",params![scope.origin,scope.account,scope.workspace,operation.id]).map_err(|error|error.to_string())?;
            if operation.kind == "attachment.upload" {
                db.execute("DELETE FROM hosted_files WHERE origin=? AND account=? AND workspace=? AND id=? AND uploaded=0",params![scope.origin,scope.account,scope.workspace,operation.payload["attachmentId"].as_str()]).map_err(|error|error.to_string())?;
            }
            db.execute(
                "INSERT OR IGNORE INTO hosted_discarded VALUES(?,?,?,?)",
                params![scope.origin, scope.account, scope.workspace, operation.id],
            )
            .map_err(|error| error.to_string())?;
            // Keep input receipts: a stale editor cannot resurrect a discarded operation.
        }
        "retry" => {
            if request.payload.is_some() || request.base_version.is_some() {
                if operation.kind == "attachment.upload" {
                    return Err("Retained upload contents cannot be replaced. Discard this upload and choose the file again.".into());
                }
                if operation.status == "pending" {
                    return Err(
                        "Reconcile the original request before changing its contents".into(),
                    );
                }
                if operation.status == "conflict"
                    && (request.base_version.is_none()
                        || request.base_version != operation.server_version)
                {
                    return Err(
                        "Review the current server version before replacing this conflict".into(),
                    );
                }
                let previous = operation.id.clone();
                operation.id = uuid::Uuid::new_v4().to_string();
                if let Some(payload) = &request.payload {
                    operation.payload = payload.clone();
                }
                if let Some(version) = &request.base_version {
                    operation.base_version = Some(version.clone());
                }
                validate_operation(&operation)?;
                db.execute("UPDATE hosted_operations SET id=? WHERE origin=? AND account=? AND workspace=? AND id=?",params![operation.id,scope.origin,scope.account,scope.workspace,previous]).map_err(|error|error.to_string())?;
                db.execute(
                    "INSERT OR IGNORE INTO hosted_discarded VALUES(?,?,?,?)",
                    params![scope.origin, scope.account, scope.workspace, previous],
                )
                .map_err(|error| error.to_string())?;
                operation.attempted = false;
            } else if operation.status == "conflict" {
                return Err(
                    "Review the current server version before retrying this conflict".into(),
                );
            }
            operation.status = "pending".into();
            operation.error = None;
            operation.server_value = None;
            operation.server_version = None;
            save(db, scope, &operation)?;
        }
        _ => return Err("Unknown conflict resolution".into()),
    }
    Ok(())
}
