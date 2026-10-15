use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};
use std::{path::Path, sync::Mutex};

use super::model::{apply_operation, empty_workspace, now, Operation};

pub struct Store(pub Mutex<Connection>);

fn has_table(db: &Connection, table: &str) -> Result<bool, String> {
    db.query_row(
        "SELECT EXISTS(SELECT 1 FROM sqlite_schema WHERE type='table' AND name=?)",
        [table],
        |row| row.get(0),
    )
    .map_err(|error| error.to_string())
}

// A staged upload remains recoverable even after its workspace/resource is
// revoked. Successfully uploaded copies are server-owned cache at that point.
const UNREFERENCED_ATTACHMENT: &str = "NOT EXISTS(SELECT 1 FROM operations WHERE operations.account=attachments.account AND operations.workspace=attachments.workspace AND json_extract(operation,'$.payload.attachmentId')=attachments.id)";

impl Store {
    pub fn server_snapshot(&self, account: &str, workspace: &str) -> Result<Value, String> {
        self.authorize(account, workspace)?;
        let encoded: Option<String> = self
            .0
            .lock()
            .map_err(|e| e.to_string())?
            .query_row(
                "SELECT data FROM snapshots WHERE account=? AND workspace=?",
                params![account, workspace],
                |r| r.get(0),
            )
            .optional()
            .map_err(|e| e.to_string())?;
        encoded
            .map(|v| serde_json::from_str(&v).map_err(|e| e.to_string()))
            .unwrap_or_else(|| Ok(empty_workspace(workspace)))
    }
    pub fn open(path: &Path) -> Result<Self, String> {
        let connection = Connection::open(path).map_err(|e| e.to_string())?;
        let version: u32 = connection
            .pragma_query_value(None, "user_version", |row| row.get(0))
            .map_err(|e| e.to_string())?;
        if version > 1 {
            return Err(
                "This workspace was saved by a newer FortyOne version. Update the app to open it."
                    .into(),
            );
        }
        connection
            .busy_timeout(std::time::Duration::from_secs(5))
            .map_err(|e| e.to_string())?;
        connection.execute_batch("PRAGMA journal_mode=WAL; PRAGMA synchronous=FULL; PRAGMA foreign_keys=ON; PRAGMA secure_delete=ON; PRAGMA max_page_count=262144;
            CREATE TABLE IF NOT EXISTS settings(key TEXT PRIMARY KEY,value TEXT NOT NULL);
            CREATE TABLE IF NOT EXISTS accounts(id TEXT PRIMARY KEY,user TEXT NOT NULL,workspaces TEXT NOT NULL,verified INTEGER NOT NULL);
            CREATE TABLE IF NOT EXISTS snapshots(account TEXT NOT NULL,workspace TEXT NOT NULL,data TEXT NOT NULL,synced INTEGER NOT NULL,PRIMARY KEY(account,workspace));
            CREATE TABLE IF NOT EXISTS operations(id TEXT PRIMARY KEY,account TEXT NOT NULL,workspace TEXT NOT NULL,operation TEXT NOT NULL);
            CREATE INDEX IF NOT EXISTS operations_scope ON operations(account,workspace);
            CREATE TABLE IF NOT EXISTS drafts(account TEXT NOT NULL,workspace TEXT NOT NULL,resource TEXT NOT NULL,kind TEXT NOT NULL,payload TEXT NOT NULL,PRIMARY KEY(account,workspace,resource,kind));
            CREATE TABLE IF NOT EXISTS attachments(id TEXT PRIMARY KEY,account TEXT NOT NULL,workspace TEXT NOT NULL,name TEXT NOT NULL,bytes BLOB NOT NULL);
            CREATE TABLE IF NOT EXISTS workspace_settings(account TEXT NOT NULL,workspace TEXT NOT NULL,data TEXT NOT NULL,PRIMARY KEY(account,workspace));
            CREATE TABLE IF NOT EXISTS workspace_access(account TEXT NOT NULL,workspace TEXT NOT NULL,verified INTEGER NOT NULL,denied TEXT,PRIMARY KEY(account,workspace));
            CREATE TABLE IF NOT EXISTS identity_aliases(account TEXT NOT NULL,workspace TEXT NOT NULL,provisional TEXT NOT NULL,resource TEXT NOT NULL,kind TEXT NOT NULL,version TEXT,creation_hash TEXT NOT NULL,PRIMARY KEY(account,workspace,provisional));
            PRAGMA user_version=1;").map_err(|e| e.to_string())?;
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            std::fs::set_permissions(path, std::fs::Permissions::from_mode(0o600))
                .map_err(|e| e.to_string())?;
        }
        Ok(Self(Mutex::new(connection)))
    }

    pub fn setting(&self, key: &str) -> Result<Option<Value>, String> {
        let db = self.0.lock().map_err(|e| e.to_string())?;
        let value: Option<String> = db
            .query_row("SELECT value FROM settings WHERE key=?", [key], |r| {
                r.get(0)
            })
            .optional()
            .map_err(|e| e.to_string())?;
        value
            .map(|v| serde_json::from_str(&v).map_err(|e| e.to_string()))
            .transpose()
    }

    pub fn set_setting(&self, key: &str, value: &Value) -> Result<(), String> {
        self.0.lock().map_err(|e| e.to_string())?.execute("INSERT INTO settings VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", params![key,value.to_string()]).map_err(|e| e.to_string())?;
        Ok(())
    }

    pub fn account(&self) -> Result<String, String> {
        self.setting("account")?
            .and_then(|v| v.as_str().map(str::to_owned))
            .filter(|v| !v.is_empty())
            .ok_or_else(|| "Sign in to continue".into())
    }

    pub fn authorize(&self, account: &str, workspace: &str) -> Result<Value, String> {
        let membership = self.membership(account, workspace)?;
        let access: Option<(u64, Option<String>)> = self
            .0
            .lock()
            .map_err(|e| e.to_string())?
            .query_row(
                "SELECT verified,denied FROM workspace_access WHERE account=? AND workspace=?",
                params![account, workspace],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )
            .optional()
            .map_err(|e| e.to_string())?;
        if let Some((verified, denied)) = access {
            if let Some(reason) = denied {
                return Err(reason);
            }
            if now().saturating_sub(verified) > 7 * 24 * 3600 {
                return Err("Reconnect to renew this workspace's offline access. Your saved work is retained.".into());
            }
        }
        Ok(membership)
    }

    pub fn grant_workspace(&self, account: &str, workspace: &str) -> Result<(), String> {
        self.0.lock().map_err(|e|e.to_string())?.execute("INSERT INTO workspace_access VALUES(?,?,?,NULL) ON CONFLICT(account,workspace) DO UPDATE SET verified=excluded.verified,denied=NULL",params![account,workspace,now()]).map_err(|e|e.to_string())?;
        Ok(())
    }

    pub fn deny_workspace(&self, account: &str, workspace: &str) -> Result<(), String> {
        let mut db = self.0.lock().map_err(|e| e.to_string())?;
        let tx = db.transaction().map_err(|e| e.to_string())?;
        tx.execute("INSERT INTO workspace_access VALUES(?,?,0,?) ON CONFLICT(account,workspace) DO UPDATE SET denied=excluded.denied",params![account,workspace,"Reconnect and verify access to this workspace. Unsynced work is retained for export."]).map_err(|e|e.to_string())?;
        for table in [
            "snapshots",
            "workspace_settings",
            "cached_files",
            "notification_receipts",
        ] {
            if !has_table(&tx, table)? {
                continue;
            }
            tx.execute(
                &format!("DELETE FROM {table} WHERE account=? AND workspace=?"),
                params![account, workspace],
            )
            .map_err(|e| e.to_string())?;
        }
        tx.execute(&format!("DELETE FROM attachments WHERE account=? AND workspace=? AND {UNREFERENCED_ATTACHMENT}"),params![account,workspace]).map_err(|e|e.to_string())?;
        tx.execute("DELETE FROM drafts WHERE account=? AND workspace=? AND kind='collaboration' AND json_extract(payload,'$.dirty')=0",params![account,workspace]).map_err(|e|e.to_string())?;
        tx.commit().map_err(|e| e.to_string())
    }

    pub fn membership(&self, account: &str, workspace: &str) -> Result<Value, String> {
        let db = self.0.lock().map_err(|e| e.to_string())?;
        let (workspaces, verified): (String, u64) = db
            .query_row(
                "SELECT workspaces,verified FROM accounts WHERE id=?",
                [account],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )
            .map_err(|_| "Sign in to continue".to_owned())?;
        if now().saturating_sub(verified) > 7 * 24 * 3600 {
            return Err(
                "Reconnect to renew offline access. Your saved changes are retained.".into(),
            );
        }
        let values: Vec<Value> = serde_json::from_str(&workspaces).map_err(|e| e.to_string())?;
        values
            .into_iter()
            .find(|v| v["id"].as_str() == Some(workspace))
            .ok_or_else(|| "You do not have access to this workspace".into())
    }

    pub fn save_account(&self, user: &Value, workspaces: &Value) -> Result<(), String> {
        let id = user["id"].as_str().ok_or("Invalid account response")?;
        if !workspaces.is_array() {
            return Err("Invalid workspace response".into());
        }
        let mut db = self.0.lock().map_err(|e| e.to_string())?;
        let tx = db.transaction().map_err(|e| e.to_string())?;
        tx.execute("INSERT INTO accounts VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET user=excluded.user,workspaces=excluded.workspaces,verified=excluded.verified",params![id,user.to_string(),workspaces.to_string(),now()]).map_err(|e|e.to_string())?;
        tx.execute("INSERT INTO settings VALUES('account',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value",[json!(id).to_string()]).map_err(|e|e.to_string())?;
        // Revoked membership removes server-owned cached data immediately.
        // Unsynced user work remains exportable, but is never exposed as an
        // accessible workspace or sent with another account's session.
        for table in [
            "snapshots",
            "workspace_settings",
            "cached_files",
            "notification_receipts",
        ] {
            if !has_table(&tx, table)? {
                continue;
            }
            tx.execute(&format!("DELETE FROM {table} WHERE account=? AND workspace NOT IN (SELECT json_extract(value,'$.id') FROM json_each(?))"),params![id,workspaces.to_string()]).map_err(|e|e.to_string())?;
        }
        tx.execute(&format!("DELETE FROM attachments WHERE account=? AND workspace NOT IN (SELECT json_extract(value,'$.id') FROM json_each(?)) AND {UNREFERENCED_ATTACHMENT}"),params![id,workspaces.to_string()]).map_err(|e|e.to_string())?;
        tx.execute("DELETE FROM drafts WHERE account=? AND workspace NOT IN (SELECT json_extract(value,'$.id') FROM json_each(?)) AND kind='collaboration' AND json_extract(payload,'$.dirty')=0",params![id,workspaces.to_string()]).map_err(|e|e.to_string())?;
        tx.commit().map_err(|e| e.to_string())
    }

    pub fn bootstrap(&self) -> Result<Value, String> {
        let account = self.account().ok();
        let active = self.setting("workspace")?;
        let db = self.0.lock().map_err(|e| e.to_string())?;
        let row: Option<(String, String, u64)> = account
            .as_ref()
            .map(|id| {
                db.query_row(
                    "SELECT user,workspaces,verified FROM accounts WHERE id=?",
                    [id],
                    |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)),
                )
                .optional()
            })
            .transpose()
            .map_err(|e| e.to_string())?
            .flatten();
        match row {
            Some((user, workspaces, verified)) => Ok(
                json!({"user":serde_json::from_str::<Value>(&user).map_err(|e|e.to_string())?,"workspaces":serde_json::from_str::<Value>(&workspaces).map_err(|e|e.to_string())?,"activeWorkspaceId":active,"needsSignIn":now().saturating_sub(verified)>7*24*3600}),
            ),
            None => {
                Ok(json!({"user":null,"workspaces":[],"activeWorkspaceId":null,"needsSignIn":true}))
            }
        }
    }

    pub fn operations(&self, account: &str, workspace: &str) -> Result<Vec<Operation>, String> {
        let db = self.0.lock().map_err(|e| e.to_string())?;
        let mut query = db
            .prepare(
                "SELECT operation FROM operations WHERE account=? AND workspace=? ORDER BY rowid",
            )
            .map_err(|e| e.to_string())?;
        let values = query
            .query_map(params![account, workspace], |r| r.get::<_, String>(0))
            .map_err(|e| e.to_string())?;
        values
            .map(|v| {
                serde_json::from_str(&v.map_err(|e| e.to_string())?).map_err(|e| e.to_string())
            })
            .collect()
    }

    pub fn workspace(&self, account: &str, workspace: &str) -> Result<Value, String> {
        self.authorize(account, workspace)?;
        let stored: Option<(String, u64)> = self
            .0
            .lock()
            .map_err(|e| e.to_string())?
            .query_row(
                "SELECT data,synced FROM snapshots WHERE account=? AND workspace=?",
                params![account, workspace],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )
            .optional()
            .map_err(|e| e.to_string())?;
        let mut value = match stored {
            Some((data, _)) => {
                let mut v: Value = serde_json::from_str(&data).map_err(|e| e.to_string())?;
                v["lastSyncedAt"] = v.get("syncedAt").cloned().unwrap_or(Value::Null);
                v
            }
            None => empty_workspace(workspace),
        };
        let operations = self.operations(account, workspace)?;
        for op in &operations {
            apply_operation(&mut value, op);
        }
        let drafts = {
            let db = self.0.lock().map_err(|e| e.to_string())?;
            let mut query=db.prepare("SELECT resource,payload FROM drafts WHERE account=? AND workspace=? AND kind='collaboration' AND json_extract(payload,'$.dirty')=1").map_err(|e|e.to_string())?;
            let rows = query
                .query_map(params![account, workspace], |r| {
                    Ok((r.get::<_, String>(0)?, r.get::<_, String>(1)?))
                })
                .map_err(|e| e.to_string())?;
            rows.collect::<Result<Vec<_>, _>>()
                .map_err(|e| e.to_string())?
        };
        for (id, encoded) in drafts {
            let draft: Value = serde_json::from_str(&encoded).map_err(|e| e.to_string())?;
            if let Some(document) = value["documents"].as_array_mut().and_then(|items| {
                items
                    .iter_mut()
                    .find(|item| item["id"].as_str() == Some(&id))
            }) {
                if draft["epoch"] == document["collaborationEpoch"] {
                    for field in ["title", "contentHtml", "contentText"] {
                        if draft[field].is_string() {
                            document[field] = draft[field].clone();
                        }
                    }
                    document["pendingSync"] = json!(true);
                }
            }
        }
        value["pending"] = serde_json::to_value(operations).map_err(|e| e.to_string())?;
        value["workspaceId"] = json!(workspace);
        value["downloadState"] = self
            .setting(&super::download::state_key(account, workspace))?
            .unwrap_or(Value::Null);
        Ok(value)
    }

    #[cfg(test)]
    pub fn enqueue(
        &self,
        account: &str,
        workspace: &str,
        operation: &Operation,
    ) -> Result<(), String> {
        self.enqueue_consuming(account, workspace, operation, None)
    }

    pub fn mark_operation(&self, account: &str, operation: &Operation) -> Result<(), String> {
        self.0
            .lock()
            .map_err(|e| e.to_string())?
            .execute(
                "UPDATE operations SET operation=? WHERE id=? AND account=?",
                params![
                    serde_json::to_string(operation).map_err(|e| e.to_string())?,
                    operation.id,
                    account
                ],
            )
            .map_err(|e| e.to_string())?;
        Ok(())
    }

    pub fn snapshot(&self, account: &str, workspace: &str, data: &Value) -> Result<(), String> {
        self.authorize(account, workspace)?;
        for key in ["stories", "documents", "comments", "teams", "statuses"] {
            if !data[key].is_array() {
                return Err(format!("Incomplete workspace response: {key}"));
            }
        }
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
        let mut merged = previous
            .map(|v| serde_json::from_str::<Value>(&v))
            .transpose()
            .map_err(|e| e.to_string())?
            .unwrap_or_else(|| empty_workspace(workspace));
        for key in ["stories", "documents", "comments", "teams", "statuses"] {
            let incoming = data[key].as_array().ok_or("Invalid snapshot collection")?;
            let complete = data["complete"][key].as_bool() == Some(true);
            let removed = data["removed"][key].as_array().cloned().unwrap_or_default();
            let mut items = if complete {
                Vec::new()
            } else {
                merged[key].as_array().cloned().unwrap_or_default()
            };
            items.retain(|item| {
                !removed.iter().any(|id| id == &item["id"])
                    && !incoming.iter().any(|new| new["id"] == item["id"])
            });
            if key == "comments" {
                let completed = data["commentsCompleteFor"]
                    .as_array()
                    .cloned()
                    .unwrap_or_default();
                items.retain(|item| !completed.iter().any(|id| id == &item["storyId"]));
            }
            items.extend(incoming.clone());
            merged[key] = Value::Array(items);
        }
        let removed_stories = data["removed"]["stories"]
            .as_array()
            .cloned()
            .unwrap_or_default();
        if let Some(comments) = merged["comments"].as_array_mut() {
            comments.retain(|comment| !removed_stories.contains(&comment["storyId"]));
        }
        let incoming_complete = data["commentsCompleteFor"]
            .as_array()
            .cloned()
            .unwrap_or_default();
        let mut complete_threads = merged["commentsCompleteFor"]
            .as_array()
            .cloned()
            .unwrap_or_default();
        complete_threads.retain(|id| {
            !removed_stories.contains(id)
                && !data["stories"]
                    .as_array()
                    .into_iter()
                    .flatten()
                    .any(|story| &story["id"] == id && !incoming_complete.contains(id))
        });
        for id in incoming_complete {
            if !removed_stories.contains(&id) && !complete_threads.contains(&id) {
                complete_threads.push(id);
            }
        }
        merged["commentsCompleteFor"] = json!(complete_threads);
        let has_media = has_table(&tx, "cached_files")?;
        for (collection, kind) in [("stories", "story"), ("documents", "document")] {
            for id in data["removed"][collection]
                .as_array()
                .into_iter()
                .flatten()
                .filter_map(Value::as_str)
            {
                if has_media {
                    tx.execute("DELETE FROM cached_files WHERE account=? AND workspace=? AND resource=? AND kind=?",params![account,workspace,id,kind]).map_err(|e|e.to_string())?;
                }
                if let Some(attachments) = merged["attachments"].as_array_mut() {
                    for attachment in attachments.iter().filter(|attachment| {
                        attachment["resourceId"].as_str() == Some(id)
                            && attachment["targetKind"].as_str() == Some(kind)
                    }) {
                        if let Some(attachment_id) = attachment["id"].as_str() {
                            tx.execute(&format!("DELETE FROM attachments WHERE account=? AND workspace=? AND id=? AND {UNREFERENCED_ATTACHMENT}"),params![account,workspace,attachment_id]).map_err(|e|e.to_string())?;
                        }
                    }
                    attachments.retain(|attachment| {
                        attachment["resourceId"].as_str() != Some(id)
                            || attachment["targetKind"].as_str() != Some(kind)
                    });
                }
            }
        }
        for id in data["removed"]["documents"]
            .as_array()
            .into_iter()
            .flatten()
            .filter_map(Value::as_str)
        {
            tx.execute("DELETE FROM drafts WHERE account=? AND workspace=? AND resource=? AND kind='collaboration' AND json_extract(payload,'$.dirty')=0",params![account,workspace,id]).map_err(|e|e.to_string())?;
        }
        for (key, complete) in data["complete"].as_object().into_iter().flatten() {
            merged["complete"][key] = complete.clone();
        }
        merged["syncedAt"] = data["syncedAt"].clone();
        tx.execute("INSERT INTO snapshots VALUES(?,?,?,?) ON CONFLICT(account,workspace) DO UPDATE SET data=excluded.data,synced=excluded.synced",params![account,workspace,merged.to_string(),now()]).map_err(|e|e.to_string())?;
        tx.commit().map_err(|e| e.to_string())
    }
}
