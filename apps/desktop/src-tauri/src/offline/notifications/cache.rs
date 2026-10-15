use super::super::{
    model::{empty_workspace, now},
    store::Store,
    Runtime,
};
use super::{
    codec::{Notification, RealtimeEvent},
    current,
};
use rusqlite::{params, OptionalExtension};
use serde_json::{json, Value};
use tauri::{AppHandle, Emitter, Manager};

const MAX_INBOX: usize = 100;

fn verify_scope(db: &rusqlite::Connection, account: &str, workspace: &str) -> Result<(), String> {
    let valid: bool = db.query_row(
        "SELECT EXISTS(SELECT 1 FROM accounts WHERE id=? AND verified>=? AND EXISTS(SELECT 1 FROM json_each(accounts.workspaces) WHERE json_extract(value,'$.id')=?)) AND EXISTS(SELECT 1 FROM settings WHERE key='account' AND value=?) AND NOT EXISTS(SELECT 1 FROM workspace_access WHERE account=? AND workspace=? AND (denied IS NOT NULL OR verified<?))",
        params![account, now().saturating_sub(7 * 24 * 3600), workspace, json!(account).to_string(), account, workspace, now().saturating_sub(7 * 24 * 3600)],
        |row| row.get(0),
    ).map_err(|error| error.to_string())?;
    if valid {
        Ok(())
    } else {
        Err("Account or workspace access changed".into())
    }
}

pub(super) fn prune_receipts(store: &Store, account: &str) -> Result<(), String> {
    store.0.lock().map_err(|error|error.to_string())?.execute(
        "DELETE FROM notification_receipts WHERE account=? AND workspace NOT IN (SELECT json_extract(value,'$.id') FROM accounts,json_each(accounts.workspaces) WHERE accounts.id=?)",
        params![account,account],
    ).map_err(|error|error.to_string())?;
    Ok(())
}

pub(super) fn initialize(store: &Store) -> Result<(), String> {
    store.0.lock().map_err(|e| e.to_string())?.execute_batch(
        "CREATE TABLE IF NOT EXISTS notification_receipts(account TEXT NOT NULL,workspace TEXT NOT NULL,id TEXT NOT NULL,version TEXT NOT NULL,PRIMARY KEY(account,workspace,id,version));"
    ).map_err(|e| e.to_string())
}

pub(super) fn update_cache(
    store: &Store,
    account: &str,
    workspace: &str,
    event: &RealtimeEvent,
) -> Result<bool, String> {
    store.authorize(account, workspace)?;
    let mut db = store.0.lock().map_err(|e| e.to_string())?;
    let tx = db.transaction().map_err(|e| e.to_string())?;
    verify_scope(&tx, account, workspace)?;
    let stored: Option<String> = tx
        .query_row(
            "SELECT data FROM snapshots WHERE account=? AND workspace=?",
            params![account, workspace],
            |r| r.get(0),
        )
        .optional()
        .map_err(|e| e.to_string())?;
    let mut snapshot = stored
        .map(|v| serde_json::from_str::<Value>(&v))
        .transpose()
        .map_err(|e| e.to_string())?
        .unwrap_or_else(|| empty_workspace(workspace));
    let mut fresh = false;
    match event {
        RealtimeEvent::Notification(notification) => {
            fresh = tx
                .execute(
                    "INSERT OR IGNORE INTO notification_receipts VALUES(?,?,?,?)",
                    params![account, workspace, notification.id, notification.created_at],
                )
                .map_err(|e| e.to_string())?
                == 1;
            let mut items = snapshot["notifications"]
                .as_array()
                .cloned()
                .unwrap_or_default();
            let previous = items
                .iter()
                .find(|item| item["id"].as_str() == Some(&notification.id));
            let unread_increment = notification.read_at.is_none()
                && previous.is_none_or(|item| !item["readAt"].is_null());
            if fresh {
                items.retain(|item| item["id"].as_str() != Some(&notification.id));
                items.insert(
                    0,
                    serde_json::to_value(notification).map_err(|e| e.to_string())?,
                );
                items.truncate(MAX_INBOX);
                snapshot["notifications"] = json!(items);
                if unread_increment {
                    snapshot["unreadNotificationCount"] = json!(snapshot
                        ["unreadNotificationCount"]
                        .as_u64()
                        .unwrap_or(0)
                        .saturating_add(1));
                }
            }
        }
        RealtimeEvent::Story(update) => {
            if let Some(story) = snapshot["stories"]
                .as_array_mut()
                .and_then(|stories| {
                    stories
                        .iter_mut()
                        .find(|story| story["id"].as_str() == Some(&update.story_id))
                })
                .and_then(Value::as_object_mut)
            {
                for (key, value) in &update.changes {
                    if key != "checklistChanged" {
                        story.insert(key.clone(), value.clone());
                    }
                }
            }
        }
        RealtimeEvent::Calendar => {}
    }
    snapshot["realtimeStale"] = json!(true);
    tx.execute("INSERT INTO snapshots VALUES(?,?,?,?) ON CONFLICT(account,workspace) DO UPDATE SET data=excluded.data",params![account,workspace,snapshot.to_string(),now()]).map_err(|e| e.to_string())?;
    tx.commit().map_err(|e| e.to_string())?;
    Ok(fresh)
}

pub(super) fn badge(app: &AppHandle, runtime: &Runtime) {
    let count = (|| -> Result<i64, String> {
        let account = runtime.store.account()?;
        let bootstrap = runtime.store.bootstrap()?;
        let mut count = 0u64;
        for workspace in bootstrap["workspaces"].as_array().into_iter().flatten() {
            if let Some(id) = workspace["id"].as_str() {
                if let Ok(snapshot) = runtime.store.workspace(&account, id) {
                    count = count
                        .saturating_add(snapshot["unreadNotificationCount"].as_u64().unwrap_or(0));
                }
            }
        }
        Ok(count.min(i64::MAX as u64) as i64)
    })()
    .unwrap_or(0);
    if let Some(window) = app.get_webview_window("main") {
        #[cfg(not(target_os = "windows"))]
        let _ = window.set_badge_count(if count > 0 { Some(count) } else { None });
        #[cfg(target_os = "windows")]
        let _ = window.set_overlay_icon(if count > 0 {
            app.default_window_icon().cloned()
        } else {
            None
        });
    }
    let _ = app.emit("desktop-unread-count", count);
}

pub(super) async fn catch_up(
    app: &AppHandle,
    runtime: &Runtime,
    account: &str,
    workspace: &str,
    slug: &str,
    token: &str,
) -> Result<(), String> {
    let inbox = runtime
        .transport
        .get(&format!("/workspaces/{slug}/notifications"), token)
        .await?;
    let unread = runtime
        .transport
        .get(
            &format!("/workspaces/{slug}/notifications/unread-count"),
            token,
        )
        .await?;
    if !current(runtime, account, workspace) {
        return Err("Account or workspace access changed".into());
    }
    let items: Vec<Value> = inbox["notifications"]
        .as_array()
        .ok_or("Invalid inbox response")?
        .iter()
        .filter_map(|item| serde_json::from_value::<Notification>(item.clone()).ok())
        .filter(|item| item.valid_for(account, workspace))
        .take(MAX_INBOX)
        .filter_map(|item| serde_json::to_value(item).ok())
        .collect();
    let count = unread
        .as_u64()
        .or_else(|| unread["count"].as_u64())
        .ok_or("Invalid unread count")?;
    let mut db = runtime.store.0.lock().map_err(|e| e.to_string())?;
    let tx = db.transaction().map_err(|e| e.to_string())?;
    verify_scope(&tx, account, workspace)?;
    let stored: Option<String> = tx
        .query_row(
            "SELECT data FROM snapshots WHERE account=? AND workspace=?",
            params![account, workspace],
            |r| r.get(0),
        )
        .optional()
        .map_err(|e| e.to_string())?;
    let mut snapshot = stored
        .map(|v| serde_json::from_str::<Value>(&v))
        .transpose()
        .map_err(|e| e.to_string())?
        .unwrap_or_else(|| empty_workspace(workspace));
    snapshot["notifications"] = json!(items);
    snapshot["unreadNotificationCount"] = json!(count);
    tx.execute("INSERT INTO snapshots VALUES(?,?,?,?) ON CONFLICT(account,workspace) DO UPDATE SET data=excluded.data",params![account,workspace,snapshot.to_string(),now()]).map_err(|e|e.to_string())?;
    tx.commit().map_err(|e| e.to_string())?;
    drop(db);
    badge(app, runtime);
    let _ = app.emit(
        "desktop-realtime",
        json!({"workspaceId":workspace,"workspaceSlug":slug,"event":{"type":"reconnected"}}),
    );
    Ok(())
}
