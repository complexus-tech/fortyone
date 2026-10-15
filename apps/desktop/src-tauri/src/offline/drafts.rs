use super::{identity, require_account, require_account_in, runtime};
use rusqlite::{params, OptionalExtension};
use serde_json::{json, Value};
use tauri::WebviewWindow;

fn validate(kind: &str, resource: &str) -> Result<(), String> {
    if !matches!(
        kind,
        "story" | "document" | "comment" | "yjs" | "collaboration"
    ) || resource.len() > 128
        || resource.is_empty()
    {
        return Err("Invalid draft".into());
    }
    Ok(())
}

#[tauri::command]
pub fn desktop_pending_collaborations(
    window: WebviewWindow,
    expected_account_id: String,
) -> Result<Vec<Value>, String> {
    let runtime = runtime(&window)?;
    let account = require_account(&runtime, &expected_account_id)?;
    let rows = {
        let db = runtime.store.0.lock().map_err(|e| e.to_string())?;
        let mut query = db.prepare("SELECT workspace,resource,payload FROM drafts WHERE account=? AND kind='collaboration' AND json_extract(payload,'$.dirty')=1").map_err(|e|e.to_string())?;
        let values = query
            .query_map([&account], |r| {
                Ok((
                    r.get::<_, String>(0)?,
                    r.get::<_, String>(1)?,
                    r.get::<_, String>(2)?,
                ))
            })
            .map_err(|e| e.to_string())?;
        values
            .collect::<Result<Vec<_>, _>>()
            .map_err(|e| e.to_string())?
    };
    let mut result = Vec::new();
    for (workspace, document, payload) in rows {
        if runtime.store.authorize(&account, &workspace).is_ok() {
            result.push(json!({"workspaceId":workspace,"documentId":document,"payload":serde_json::from_str::<Value>(&payload).map_err(|e|e.to_string())?}));
        }
    }
    Ok(result)
}

#[tauri::command]
pub fn desktop_save_draft(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    resource_id: String,
    kind: String,
    payload: Value,
) -> Result<(), String> {
    let runtime = runtime(&window)?;
    let account = require_account(&runtime, &expected_account_id)?;
    let db = runtime.store.0.lock().map_err(|e| e.to_string())?;
    retain_draft(&db, &account, &workspace_id, &resource_id, &kind, &payload)
}

/// Preserve edits already held by this account even when its offline lease
/// expires or workspace access is revoked. This is a write-only recovery path;
/// it never renews permissions or grants access to downloaded server data.
fn retain_draft(
    db: &rusqlite::Connection,
    account: &str,
    workspace: &str,
    resource: &str,
    kind: &str,
    payload: &Value,
) -> Result<(), String> {
    require_account_in(db, account)?;
    validate(kind, resource)?;
    let limit = if kind == "collaboration" {
        12 * 1024 * 1024
    } else {
        4_000_000
    };
    if !payload.is_object() || payload.to_string().len() > limit {
        return Err("This draft exceeds the local save limit".into());
    }
    let known: bool = db.query_row(
        "SELECT EXISTS(SELECT 1 FROM accounts WHERE id=?1 AND (EXISTS(SELECT 1 FROM json_each(accounts.workspaces) WHERE json_extract(value,'$.id')=?2) OR EXISTS(SELECT 1 FROM workspace_access WHERE account=?1 AND workspace=?2)))",
        params![account, workspace], |row| row.get(0),
    ).map_err(|error| error.to_string())?;
    if !known {
        return Err(
            "This workspace is not known to your account. Your editor has not been cleared.".into(),
        );
    }
    if matches!(kind, "collaboration" | "yjs") && payload["dirty"] != true {
        let earliest = super::model::now().saturating_sub(7 * 24 * 3600);
        let readable: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM accounts WHERE id=?1 AND verified>=?3 AND EXISTS(SELECT 1 FROM json_each(accounts.workspaces) WHERE json_extract(value,'$.id')=?2)) AND NOT EXISTS(SELECT 1 FROM workspace_access WHERE account=?1 AND workspace=?2 AND (denied IS NOT NULL OR verified<?3))",
            params![account,workspace,earliest], |row| row.get(0),
        ).map_err(|error| error.to_string())?;
        if !readable {
            return Err(
                "Reconnect to refresh this document. Unsynced edits can still be retained locally."
                    .into(),
            );
        }
    }
    let (resource, payload) =
        identity::normalize_draft(db, account, workspace, resource, kind, payload)?;
    db.execute("INSERT INTO drafts VALUES(?,?,?,?,?) ON CONFLICT(account,workspace,resource,kind) DO UPDATE SET payload=excluded.payload",params![account,workspace,resource,kind,payload.to_string()]).map_err(|_|"Could not save this draft on your device. Check available disk space before leaving.".to_owned())?;
    Ok(())
}

#[tauri::command]
pub fn desktop_read_draft(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    resource_id: String,
    kind: String,
) -> Result<Option<Value>, String> {
    let runtime = runtime(&window)?;
    let account = require_account(&runtime, &expected_account_id)?;
    runtime.store.authorize(&account, &workspace_id)?;
    validate(&kind, &resource_id)?;
    let db = runtime.store.0.lock().map_err(|e| e.to_string())?;
    require_account_in(&db, &account)?;
    let resource_id = identity::resolve_resource(&db, &account, &workspace_id, &resource_id)?;
    let value: Option<String> = db
        .query_row(
            "SELECT payload FROM drafts WHERE account=? AND workspace=? AND resource=? AND kind=?",
            params![account, workspace_id, resource_id, kind],
            |r| r.get(0),
        )
        .optional()
        .map_err(|e| e.to_string())?;
    value
        .map(|v| {
            serde_json::from_str::<Value>(&v)
                .map(|p| json!({"payload":p}))
                .map_err(|e| e.to_string())
        })
        .transpose()
}

#[tauri::command]
pub fn desktop_delete_draft(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    resource_id: String,
    kind: String,
) -> Result<(), String> {
    let runtime = runtime(&window)?;
    let account = require_account(&runtime, &expected_account_id)?;
    runtime.store.authorize(&account, &workspace_id)?;
    validate(&kind, &resource_id)?;
    let db = runtime.store.0.lock().map_err(|e| e.to_string())?;
    require_account_in(&db, &account)?;
    let resource_id = identity::resolve_resource(&db, &account, &workspace_id, &resource_id)?;
    db.execute(
        "DELETE FROM drafts WHERE account=? AND workspace=? AND resource=? AND kind=?",
        params![account, workspace_id, resource_id, kind],
    )
    .map_err(|e| e.to_string())?;
    Ok(())
}

/// A background upload may finish after a newer edit has been saved. Only the
/// exact dirty state that was uploaded can be replaced with its clean merge.
fn acknowledge_collaboration(
    db: &rusqlite::Connection,
    account: &str,
    workspace: &str,
    document: &str,
    expected_state: &str,
    expected_name: &str,
    payload: &Value,
) -> Result<bool, String> {
    require_account_in(db, account)?;
    if !payload.is_object()
        || payload["dirty"] != false
        || payload["name"].as_str() != Some(expected_name)
        || payload["state"].as_str().is_none()
        || payload.to_string().len() > 12 * 1024 * 1024
    {
        return Err("Invalid acknowledged document state".into());
    }
    let changed = db.execute(
        "UPDATE drafts SET payload=? WHERE account=? AND workspace=? AND resource=? AND kind='collaboration' AND json_extract(payload,'$.dirty')=1 AND json_extract(payload,'$.state')=? AND json_extract(payload,'$.name')=?",
        params![payload.to_string(), account, workspace, document, expected_state, expected_name],
    ).map_err(|error| error.to_string())?;
    Ok(changed == 1)
}

#[tauri::command]
pub fn desktop_ack_collaboration(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    document_id: String,
    expected_state: String,
    expected_name: String,
    payload: Value,
) -> Result<bool, String> {
    let runtime = runtime(&window)?;
    let account = require_account(&runtime, &expected_account_id)?;
    runtime.store.authorize(&account, &workspace_id)?;
    validate("collaboration", &document_id)?;
    let db = runtime.store.0.lock().map_err(|error| error.to_string())?;
    let document = identity::resolve_resource(&db, &account, &workspace_id, &document_id)?;
    acknowledge_collaboration(
        &db,
        &account,
        &workspace_id,
        &document,
        &expected_state,
        &expected_name,
        &payload,
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    fn database() -> rusqlite::Connection {
        let db = rusqlite::Connection::open_in_memory().unwrap();
        db.execute_batch("CREATE TABLE settings(key TEXT PRIMARY KEY,value TEXT); CREATE TABLE drafts(account TEXT,workspace TEXT,resource TEXT,kind TEXT,payload TEXT,PRIMARY KEY(account,workspace,resource,kind)); CREATE TABLE accounts(id TEXT PRIMARY KEY,user TEXT,workspaces TEXT,verified INTEGER); CREATE TABLE workspace_access(account TEXT,workspace TEXT,verified INTEGER,denied TEXT,PRIMARY KEY(account,workspace)); CREATE TABLE identity_aliases(account TEXT,workspace TEXT,provisional TEXT,resource TEXT,kind TEXT,version TEXT,creation_hash TEXT);").unwrap();
        db.execute(
            "INSERT INTO settings VALUES('account',?)",
            [json!("account-a").to_string()],
        )
        .unwrap();
        db.execute(
            "INSERT INTO accounts VALUES('account-a','{}',?,?)",
            params![
                json!([{"id":"workspace"}]).to_string(),
                super::super::model::now()
            ],
        )
        .unwrap();
        db.execute(
            "INSERT INTO workspace_access VALUES('account-a','workspace',?,NULL)",
            [super::super::model::now()],
        )
        .unwrap();
        db.execute(
            "INSERT INTO drafts VALUES('account-a','workspace','document','collaboration',?)",
            [json!({"dirty":true,"name":"document:1","state":"newer-edit"}).to_string()],
        )
        .unwrap();
        db
    }

    #[test]
    fn expired_leases_retain_local_edits_without_restoring_read_access() {
        for expired_table in ["accounts", "workspace_access"] {
            let store = super::super::store::Store(std::sync::Mutex::new(database()));
            {
                let db = store.0.lock().unwrap();
                db.execute(&format!("UPDATE {expired_table} SET verified=0"), [])
                    .unwrap();
                retain_draft(
                    &db,
                    "account-a",
                    "workspace",
                    "story",
                    "story",
                    &json!({"title":"retained edit"}),
                )
                .unwrap();
                retain_draft(
                    &db,
                    "account-a",
                    "workspace",
                    "document",
                    "collaboration",
                    &json!({"dirty":true,"name":"document:1","state":"retained-edit"}),
                )
                .unwrap();
                let saved: String = db
                    .query_row(
                        "SELECT payload FROM drafts WHERE resource='story'",
                        [],
                        |row| row.get(0),
                    )
                    .unwrap();
                assert_eq!(
                    serde_json::from_str::<Value>(&saved).unwrap()["title"],
                    "retained edit"
                );
                let verified: u64 = db
                    .query_row(
                        &format!("SELECT verified FROM {expired_table}"),
                        [],
                        |row| row.get(0),
                    )
                    .unwrap();
                assert_eq!(verified, 0, "local retention must not renew the lease");
            }
            assert!(store.authorize("account-a", "workspace").is_err());
            assert!(store.workspace("account-a", "workspace").is_err());
        }
    }

    #[test]
    fn revoked_workspace_retains_dirty_work_without_repopulating_clean_document_cache() {
        let store = super::super::store::Store(std::sync::Mutex::new(database()));
        {
            let db = store.0.lock().unwrap();
            db.execute("UPDATE accounts SET workspaces='[]'", [])
                .unwrap();
            db.execute("UPDATE workspace_access SET denied='access revoked'", [])
                .unwrap();
            retain_draft(
                &db,
                "account-a",
                "workspace",
                "story",
                "story",
                &json!({"title":"unsynced edit"}),
            )
            .unwrap();
            retain_draft(
                &db,
                "account-a",
                "workspace",
                "document",
                "collaboration",
                &json!({"dirty":true,"state":"unsynced"}),
            )
            .unwrap();
            assert!(retain_draft(
                &db,
                "account-a",
                "workspace",
                "document",
                "collaboration",
                &json!({"dirty":false,"state":"server-cache"})
            )
            .is_err());
            let saved: String = db
                .query_row(
                    "SELECT payload FROM drafts WHERE resource='document'",
                    [],
                    |row| row.get(0),
                )
                .unwrap();
            assert_eq!(
                serde_json::from_str::<Value>(&saved).unwrap()["state"],
                "unsynced"
            );
        }
        assert!(store.authorize("account-a", "workspace").is_err());
        assert!(store.workspace("account-a", "workspace").is_err());
    }

    #[test]
    fn local_retention_rejects_unknown_and_other_accounts_scopes() {
        let db = database();
        db.execute(
            "INSERT INTO accounts VALUES('account-b','{}','[{\"id\":\"other-workspace\"}]',0)",
            [],
        )
        .unwrap();
        db.execute(
            "INSERT INTO workspace_access VALUES('account-b','other-workspace',0,NULL)",
            [],
        )
        .unwrap();
        for (account, workspace) in [
            ("account-a", "unknown"),
            ("account-a", "other-workspace"),
            ("account-b", "other-workspace"),
        ] {
            assert!(retain_draft(
                &db,
                account,
                workspace,
                "story",
                "story",
                &json!({"title":"must not save"})
            )
            .is_err());
        }
        let count: i64 = db
            .query_row(
                "SELECT count(*) FROM drafts WHERE kind='story'",
                [],
                |row| row.get(0),
            )
            .unwrap();
        assert_eq!(count, 0);
    }

    #[test]
    fn background_ack_does_not_overwrite_newer_saved_edits_or_another_epoch() {
        let db = database();
        let merged = json!({"dirty":false,"name":"document:1","state":"merged-old-edit"});
        assert!(!acknowledge_collaboration(
            &db,
            "account-a",
            "workspace",
            "document",
            "old-edit",
            "document:1",
            &merged
        )
        .unwrap());
        assert!(!acknowledge_collaboration(
            &db,
            "account-a",
            "workspace",
            "document",
            "newer-edit",
            "document:0",
            &json!({"dirty":false,"name":"document:0","state":"merged"})
        )
        .unwrap());
        let saved: String = db
            .query_row("SELECT payload FROM drafts", [], |row| row.get(0))
            .unwrap();
        assert_eq!(
            serde_json::from_str::<Value>(&saved).unwrap()["state"],
            "newer-edit"
        );
        assert!(acknowledge_collaboration(
            &db,
            "account-a",
            "workspace",
            "document",
            "newer-edit",
            "document:1",
            &merged
        )
        .unwrap());
        assert!(!acknowledge_collaboration(
            &db,
            "account-a",
            "workspace",
            "document",
            "newer-edit",
            "document:1",
            &merged
        )
        .unwrap());
    }

    #[test]
    fn stale_account_cannot_ack_a_draft() {
        let db = database();
        assert!(acknowledge_collaboration(
            &db,
            "account-b",
            "workspace",
            "document",
            "newer-edit",
            "document:1",
            &json!({"dirty":false,"name":"document:1","state":"merged"})
        )
        .is_err());
        db.execute("DELETE FROM settings", []).unwrap();
        assert!(require_account_in(&db, "account-a").is_err());
    }
}
