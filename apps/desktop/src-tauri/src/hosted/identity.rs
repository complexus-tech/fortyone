//! Provisional draft identities follow creation receipts without overwriting a
//! second editor's recovery copy or rebasing an independently stale version.
use super::*;
use serde_json::json;

pub(super) fn initialize(db: &Connection) -> Result<(), String> {
    db.execute_batch("CREATE TABLE IF NOT EXISTS hosted_draft_owners(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,key TEXT NOT NULL,source TEXT NOT NULL,PRIMARY KEY(origin,account,workspace,key));").map_err(|error|error.to_string())
}

pub(super) fn normalize(
    db: &Connection,
    scope: &BoundScope,
    key: &str,
    payload: &Value,
) -> Result<(String, Value), String> {
    super::drafts::validate_key(key)?;
    let mut parts: Vec<String> = key.split(':').map(str::to_owned).collect();
    if parts[0] != "new-story" {
        parts[1] = super::outbox::store::resolve_resource(db, scope, &parts[1])?;
        if parts.len() == 4 {
            parts[3] = super::outbox::store::resolve_resource(db, scope, &parts[3])?;
        }
    }
    let mut value = payload.clone();
    if let Some(object) = value.as_object_mut() {
        for field in ["resourceId", "storyId", "documentId", "parentId"] {
            if let Some(id) = object.get(field).and_then(Value::as_str) {
                let mapped = super::outbox::store::resolve_resource(db, scope, id)?;
                object.insert(field.into(), json!(mapped));
            }
        }
        if let Some(id) = object
            .get("baseVersion")
            .and_then(Value::as_str)
            .and_then(|value| value.strip_prefix("pending-create:"))
        {
            let version:Option<String>=db.query_row("SELECT version FROM hosted_aliases WHERE origin=? AND account=? AND workspace=? AND provisional=? AND kind='story.create'",params![scope.origin,scope.account,scope.workspace,id],|row|row.get(0)).optional().map_err(|error|error.to_string())?.flatten();
            if let Some(version) = version {
                object.insert("baseVersion".into(), json!(version));
            }
        }
    }
    Ok((parts.join(":"), value))
}

pub(super) fn ensure_writer(
    db: &Connection,
    scope: &BoundScope,
    source: &str,
    key: &str,
) -> Result<(), String> {
    if source == key {
        return Ok(());
    }
    let owner:Option<String>=db.query_row("SELECT COALESCE((SELECT source FROM hosted_draft_owners o WHERE o.origin=d.origin AND o.account=d.account AND o.workspace=d.workspace AND o.key=d.key),d.key) FROM hosted_drafts d WHERE origin=? AND account=? AND workspace=? AND key=?",params![scope.origin,scope.account,scope.workspace,key],|row|row.get(0)).optional().map_err(|error|error.to_string())?;
    if owner.is_some_and(|owner| owner != source) {
        return Err("A newer editor has a saved draft for this story. Reopen the saved story before replacing it; your current changes remain available for recovery.".into());
    }
    Ok(())
}

pub(super) fn remember_writer(
    db: &Connection,
    scope: &BoundScope,
    source: &str,
    key: &str,
) -> Result<(), String> {
    db.execute("INSERT INTO hosted_draft_owners VALUES(?,?,?,?,?) ON CONFLICT(origin,account,workspace,key) DO UPDATE SET source=excluded.source",params![scope.origin,scope.account,scope.workspace,key,source]).map_err(|error|error.to_string())?;
    Ok(())
}

pub(super) fn migrate(db: &Connection, scope: &BoundScope) -> Result<(), String> {
    let drafts: Vec<(String, String, u64)> = {
        let mut query=db.prepare("SELECT key,payload,saved FROM hosted_drafts WHERE origin=? AND account=? AND workspace=?").map_err(|error|error.to_string())?;
        let rows = query
            .query_map(
                params![scope.origin, scope.account, scope.workspace],
                |row| Ok((row.get(0)?, row.get(1)?, row.get(2)?)),
            )
            .map_err(|error| error.to_string())?;
        rows.collect::<Result<_, _>>()
            .map_err(|error| error.to_string())?
    };
    for (source, encoded, saved) in drafts {
        let payload: Value =
            serde_json::from_str(&encoded).map_err(|_| "A retained draft is invalid")?;
        let (key, payload) = normalize(db, scope, &source, &payload)?;
        if key == source {
            if payload.to_string() != encoded {
                db.execute("UPDATE hosted_drafts SET payload=? WHERE origin=? AND account=? AND workspace=? AND key=?",params![payload.to_string(),scope.origin,scope.account,scope.workspace,key]).map_err(|error|error.to_string())?;
            }
            continue;
        }
        let existing:Option<String>=db.query_row("SELECT payload FROM hosted_drafts WHERE origin=? AND account=? AND workspace=? AND key=?",params![scope.origin,scope.account,scope.workspace,key],|row|row.get(0)).optional().map_err(|error|error.to_string())?;
        if let Some(existing) = existing {
            if serde_json::from_str::<Value>(&existing)
                .map_err(|_| "A retained draft is invalid")?
                != payload
            {
                return Err("Both the local story and its saved server copy contain different recovery drafts. Review and save or discard one draft before retrying synchronization. Both copies are retained.".into());
            }
        } else {
            db.execute(
                "INSERT INTO hosted_drafts VALUES(?,?,?,?,?,?)",
                params![
                    scope.origin,
                    scope.account,
                    scope.workspace,
                    key,
                    payload.to_string(),
                    saved
                ],
            )
            .map_err(|error| error.to_string())?;
            remember_writer(db, scope, &source, &key)?;
        }
        db.execute(
            "DELETE FROM hosted_drafts WHERE origin=? AND account=? AND workspace=? AND key=?",
            params![scope.origin, scope.account, scope.workspace, source],
        )
        .map_err(|error| error.to_string())?;
        db.execute("DELETE FROM hosted_draft_owners WHERE origin=? AND account=? AND workspace=? AND key=?",params![scope.origin,scope.account,scope.workspace,source]).map_err(|error|error.to_string())?;
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    fn database() -> Connection {
        let db = Connection::open_in_memory().unwrap();
        super::super::initialize(&db).unwrap();
        db
    }
    fn draft(db: &Connection, scope: &BoundScope, key: &str, value: Value) {
        db.execute(
            "INSERT INTO hosted_drafts VALUES(?,?,?,?,?,?)",
            params![
                scope.origin,
                scope.account,
                scope.workspace,
                key,
                value.to_string(),
                now()
            ],
        )
        .unwrap();
    }
    fn alias(db: &Connection, scope: &BoundScope, local: &str, remote: &str, kind: &str) {
        db.execute(
            "INSERT INTO hosted_aliases VALUES(?,?,?,?,?,?,?,?)",
            params![
                scope.origin,
                scope.account,
                scope.workspace,
                local,
                remote,
                kind,
                "v1",
                "hash"
            ],
        )
        .unwrap();
    }
    #[test]
    fn draft_keys_and_pending_markers_follow_only_their_creation_receipt() {
        let db = database();
        let scope = super::super::tests::scope();
        let local = uuid::Uuid::new_v4().to_string();
        let remote = uuid::Uuid::new_v4().to_string();
        let comment = uuid::Uuid::new_v4().to_string();
        let saved_comment = uuid::Uuid::new_v4().to_string();
        let key = format!("story:{local}");
        let target = format!("story:{remote}");
        draft(
            &db,
            &scope,
            &key,
            json!({"title":"Newer edit","baseVersion":format!("pending-create:{local}")}),
        );
        draft(
            &db,
            &scope,
            &format!("comment:{local}:reply:{comment}"),
            json!({"html":"Reply"}),
        );
        alias(&db, &scope, &local, &remote, "story.create");
        alias(&db, &scope, &comment, &saved_comment, "comment.create");
        migrate(&db, &scope).unwrap();
        let encoded: String = db
            .query_row(
                "SELECT payload FROM hosted_drafts WHERE key=?",
                [&target],
                |row| row.get(0),
            )
            .unwrap();
        assert_eq!(
            serde_json::from_str::<Value>(&encoded).unwrap()["baseVersion"],
            "v1"
        );
        assert!(db
            .query_row(
                "SELECT 1 FROM hosted_drafts WHERE key=?",
                [format!("comment:{remote}:reply:{saved_comment}")],
                |row| row.get::<_, i32>(0)
            )
            .is_ok());
        assert_eq!(
            normalize(
                &db,
                &scope,
                &key,
                &json!({"baseVersion":"older-server-version"})
            )
            .unwrap()
            .1["baseVersion"],
            "older-server-version"
        );
        ensure_writer(&db, &scope, &key, &target).unwrap();
        remember_writer(&db, &scope, &target, &target).unwrap();
        assert!(ensure_writer(&db, &scope, &key, &target).is_err());
        let mut other = super::super::tests::scope();
        other.account = uuid::Uuid::new_v4().to_string();
        assert_eq!(normalize(&db, &other, &key, &Value::Null).unwrap().0, key);
    }
    #[test]
    fn collision_rolls_back_acknowledgement_without_overwriting_either_draft() {
        let mut db = database();
        let scope = super::super::tests::scope();
        let local = uuid::Uuid::new_v4().to_string();
        let remote = uuid::Uuid::new_v4().to_string();
        let request = super::super::outbox::EnqueueRequest {
            account_id: scope.account.clone(),
            workspace_id: scope.workspace.clone(),
            id: uuid::Uuid::new_v4().to_string(),
            kind: "story.create".into(),
            resource_id: local.clone(),
            payload: json!({"title":"Created"}),
            base_version: None,
            consumed_draft_key: None,
        };
        super::super::outbox::store::enqueue(&db, &scope, &request).unwrap();
        draft(
            &db,
            &scope,
            &format!("story:{local}"),
            json!({"title":"Local recovery"}),
        );
        draft(
            &db,
            &scope,
            &format!("story:{remote}"),
            json!({"title":"Server editor recovery"}),
        );
        let operation = super::super::outbox::store::read(&db, &scope)
            .unwrap()
            .operations
            .remove(0);
        {
            let tx = db.transaction().unwrap();
            let error = super::super::outbox::store::acknowledge(
                &tx,
                &scope,
                &operation,
                &json!({"resourceId":remote,"operationVersion":"v1"}),
            )
            .unwrap_err();
            assert!(error.contains("Both copies are retained"));
        }
        assert_eq!(
            db.query_row("SELECT COUNT(*) FROM hosted_drafts", [], |row| row
                .get::<_, u64>(0))
                .unwrap(),
            2
        );
        assert_eq!(
            db.query_row("SELECT COUNT(*) FROM hosted_aliases", [], |row| row
                .get::<_, u64>(0))
                .unwrap(),
            0
        );
        let retained = super::super::outbox::store::read(&db, &scope).unwrap();
        assert_eq!(retained.operations.len(), 1);
        assert!(retained.completed.is_empty());
    }
}
