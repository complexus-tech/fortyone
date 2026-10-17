//! Hosted pages forward only their existing live SSE events. Native code owns
//! receipt deduplication, system presentation, and validated click activation.
use rusqlite::{params, Connection};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use tauri::{AppHandle, Emitter, Manager, Url, WebviewWindow};
#[cfg(not(target_os = "macos"))]
use tauri_plugin_notification::{NotificationExt, PermissionState};

#[cfg(target_os = "macos")]
mod macos;

use super::{authorize_scope, authorize_window, ensure_scope_current};

#[derive(Clone, Deserialize, Serialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct Notification {
    account_id: String,
    workspace_id: String,
    workspace_slug: String,
    notification_id: String,
    created_at: String,
    title: String,
    body: String,
    entity_type: String,
    entity_id: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    comment_id: Option<String>,
}

impl Notification {
    fn validate(&self) -> Result<(), String> {
        for id in [
            &self.account_id,
            &self.workspace_id,
            &self.notification_id,
            &self.entity_id,
        ] {
            uuid::Uuid::parse_str(id).map_err(|_| "Invalid notification identity")?;
        }
        if let Some(comment_id) = &self.comment_id {
            let id = uuid::Uuid::parse_str(comment_id)
                .map_err(|_| "Invalid comment notification identity")?;
            if id.is_nil() || self.entity_type != "story" {
                return Err("Invalid comment notification target".into());
            }
        }
        if self.title.trim().is_empty()
            || self.title.len() > 2048
            || self.body.len() > 8192
            || self.created_at.is_empty()
            || self.created_at.len() > 64
            || !matches!(
                self.entity_type.as_str(),
                "story" | "comment" | "objective" | "key_result" | "strategy" | "document" | "chat_message"
            )
        {
            return Err("Invalid notification content".into());
        }
        Ok(())
    }
}

fn claim(db: &Connection, notification: &Notification) -> Result<bool, String> {
    db.execute(
        "INSERT OR IGNORE INTO hosted_notifications(account,workspace,id,version) VALUES(?,?,?,?)",
        params![
            notification.account_id,
            notification.workspace_id,
            notification.notification_id,
            notification.created_at
        ],
    )
    .map(|rows| rows > 0)
    .map_err(|error| error.to_string())
}

fn release_claim(db: &Connection, notification: &Notification) -> Result<(), String> {
    db.execute(
        "DELETE FROM hosted_notifications WHERE account=? AND workspace=? AND id=? AND version=?",
        params![
            notification.account_id,
            notification.workspace_id,
            notification.notification_id,
            notification.created_at
        ],
    )
    .map(|_| ())
    .map_err(|error| error.to_string())
}

fn activation_payload(notification: &Notification) -> Value {
    json!({
        "accountId":notification.account_id,"workspaceId":notification.workspace_id,
        "workspaceSlug":notification.workspace_slug,"notificationId":notification.notification_id,
        "createdAt":notification.created_at,"entityType":notification.entity_type,"entityId":notification.entity_id,
        "commentId":notification.comment_id,
        "title":"FortyOne","body":""
    })
}

pub(super) async fn request_permission(window: &WebviewWindow) -> Result<String, String> {
    #[cfg(target_os = "macos")]
    {
        let _ = window;
        macos::request_permission().await
    }
    #[cfg(not(target_os = "macos"))]
    {
        let state = window
            .notification()
            .request_permission()
            .map_err(|error| error.to_string())?;
        Ok(match state {
            PermissionState::Granted => "granted",
            PermissionState::Denied => "denied",
            _ => "prompt",
        }
        .into())
    }
}

fn clean_text(value: &str, limit: usize) -> String {
    value
        .chars()
        .filter(|character| !character.is_control())
        .take(limit)
        .collect()
}

#[tauri::command]
pub async fn hosted_deliver_notification(
    window: WebviewWindow,
    notification: Notification,
) -> Result<(), String> {
    notification.validate()?;
    let state = authorize_window(&window)?;
    let scope = authorize_scope(
        &window,
        &notification.account_id,
        &notification.workspace_id,
        Some(&notification.workspace_slug),
    )
    .await?;
    {
        let db = state.db.lock().map_err(|error| error.to_string())?;
        ensure_scope_current(&db, &scope)?;
        if !claim(&db, &notification)? || !super::preferences::read(&db)?.notifications {
            return Ok(());
        }
    }
    let result = deliver(&window, &notification, &scope).await;
    if let Err(error) = result {
        let db = state.db.lock().map_err(|error| error.to_string())?;
        release_claim(&db, &notification)?;
        return Err(format!("Could not show the system notification: {error}"));
    }
    Ok(())
}

async fn deliver(
    window: &WebviewWindow,
    notification: &Notification,
    original: &super::BoundScope,
) -> Result<(), String> {
    #[cfg(target_os = "macos")]
    let granted = macos::permission().await? == "granted";
    #[cfg(not(target_os = "macos"))]
    let granted = window
        .notification()
        .permission_state()
        .map_err(|error| error.to_string())?
        == PermissionState::Granted;
    if !granted {
        return Ok(());
    }
    // Permission callbacks can outlive navigation, logout, or a preference change.
    // Revalidate before submitting, without holding SQLite across either await.
    let current = authorize_scope(
        window,
        &notification.account_id,
        &notification.workspace_id,
        Some(&notification.workspace_slug),
    )
    .await?;
    if current.origin != original.origin || current.cookie_hash != original.cookie_hash {
        return Err("Your session changed before notification delivery".into());
    }
    {
        let state = authorize_window(window)?;
        let db = state.db.lock().map_err(|error| error.to_string())?;
        ensure_scope_current(&db, original)?;
        if !super::preferences::read(&db)?.notifications {
            return Ok(());
        }
    }
    #[cfg(target_os = "macos")]
    return macos::show(notification).await;
    #[cfg(not(target_os = "macos"))]
    {
        let id = (uuid::Uuid::parse_str(&notification.notification_id)
            .map_err(|_| "Invalid notification identity")?
            .as_u128() as u32
            & 0x7fff_ffff) as i32;
        window
            .notification()
            .builder()
            .id(id)
            .title(clean_text(&notification.title, 256))
            .body(clean_text(&notification.body, 1024))
            .extra("fortyoneHostedTarget", activation_payload(notification))
            .show()
            .map_err(|error| error.to_string())
    }
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct Badge {
    account_id: String,
    workspace_id: String,
    workspace_slug: String,
    count: u32,
}

pub fn set_badge(window: &WebviewWindow, count: u32) -> Result<(), String> {
    #[cfg(not(target_os = "windows"))]
    window
        .set_badge_count(if count > 0 {
            Some(i64::from(count))
        } else {
            None
        })
        .map_err(|error| error.to_string())?;
    #[cfg(target_os = "windows")]
    window
        .set_overlay_icon(if count > 0 {
            window.app_handle().default_window_icon().cloned()
        } else {
            None
        })
        .map_err(|error| error.to_string())?;
    window
        .emit("desktop-unread-count", count)
        .map_err(|error| error.to_string())
}

#[tauri::command]
pub async fn hosted_set_notification_badge(
    window: WebviewWindow,
    state: Badge,
) -> Result<(), String> {
    let runtime = authorize_window(&window)?;
    let scope = authorize_scope(
        &window,
        &state.account_id,
        &state.workspace_id,
        Some(&state.workspace_slug),
    )
    .await?;
    let db = runtime.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, &scope)?;
    set_badge(&window, state.count)
}

fn destination(mut current: Url, notification: &Notification, production: bool) -> Url {
    let prefix = if production {
        String::new()
    } else {
        format!("/{}", notification.workspace_slug)
    };
    current.set_query(None);
    current.set_fragment(None);
    if matches!(
        notification.entity_type.as_str(),
        "story" | "objective" | "key_result" | "strategy" | "chat_message"
    ) {
        current.set_path(&format!(
            "{prefix}/notifications/{}",
            notification.notification_id
        ));
        current
            .query_pairs_mut()
            .append_pair("entityId", &notification.entity_id)
            .append_pair("entityType", &notification.entity_type)
            .append_pair("observedCreatedAt", &notification.created_at);
        if notification.entity_type == "story" {
            if let Some(comment_id) = &notification.comment_id {
                current.query_pairs_mut().append_pair("comment", comment_id);
            }
        }
    } else {
        current.set_path(&format!("{prefix}/notifications"));
    }
    current
}

async fn activate(app: AppHandle, payload: Value) -> Result<(), String> {
    let notification: Notification =
        serde_json::from_value(payload).map_err(|_| "Invalid notification destination")?;
    notification.validate()?;
    let window = app
        .get_webview_window("main")
        .ok_or("The app window is unavailable")?;
    let state = authorize_window(&window)?;
    let scope = authorize_scope(
        &window,
        &notification.account_id,
        &notification.workspace_id,
        Some(&notification.workspace_slug),
    )
    .await?;
    let current = window.url().map_err(|error| error.to_string())?;
    let url = destination(current, &notification, cfg!(feature = "production"));
    {
        let db = state.db.lock().map_err(|error| error.to_string())?;
        ensure_scope_current(&db, &scope)?;
        window.navigate(url).map_err(|error| error.to_string())?;
    }
    window.show().map_err(|error| error.to_string())?;
    window.unminimize().map_err(|error| error.to_string())?;
    window.set_focus().map_err(|error| error.to_string())
}

pub fn install(app: &AppHandle) -> Result<(), String> {
    #[cfg(target_os = "macos")]
    return macos::install(app);
    #[cfg(not(target_os = "macos"))]
    {
        let app = app.clone();
        app.clone()
            .notification()
            .on_action(move |action| {
                if action.action_id() != "tap" {
                    return;
                }
                let Some(payload) = action
                    .notification()
                    .and_then(|notification| notification.extra().get("fortyoneHostedTarget"))
                    .cloned()
                else {
                    return;
                };
                let app = app.clone();
                tauri::async_runtime::spawn(async move {
                    if let Err(error) = activate(app.clone(), payload).await {
                        let _ = app.emit("desktop-notification-error", error);
                    }
                });
            })
            .map_err(|error| error.to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    pub(super) fn notification() -> Notification {
        Notification {
            account_id: uuid::Uuid::new_v4().to_string(),
            workspace_id: uuid::Uuid::new_v4().to_string(),
            workspace_slug: "engineering".into(),
            notification_id: uuid::Uuid::new_v4().to_string(),
            created_at: "2026-10-15T12:00:00.123456Z".into(),
            title: "New comment".into(),
            body: "Please review".into(),
            entity_type: "story".into(),
            entity_id: uuid::Uuid::new_v4().to_string(),
            comment_id: None,
        }
    }
    #[test]
    fn receipts_dedupe_per_account_workspace_and_coalesced_version() {
        let db = Connection::open_in_memory().unwrap();
        db.execute_batch("CREATE TABLE hosted_notifications(account TEXT,workspace TEXT,id TEXT,version TEXT,PRIMARY KEY(account,workspace,id,version));").unwrap();
        let mut event = notification();
        assert!(claim(&db, &event).unwrap());
        assert!(!claim(&db, &event).unwrap());
        event.created_at = "2026-10-15T12:01:00.123456Z".into();
        assert!(claim(&db, &event).unwrap());
        event.account_id = uuid::Uuid::new_v4().to_string();
        assert!(claim(&db, &event).unwrap());
    }
    #[test]
    fn failed_submission_releases_only_the_claimed_scope_and_version() {
        let db = Connection::open_in_memory().unwrap();
        db.execute_batch("CREATE TABLE hosted_notifications(account TEXT,workspace TEXT,id TEXT,version TEXT,PRIMARY KEY(account,workspace,id,version));").unwrap();
        let event = notification();
        let mut later = event.clone();
        later.created_at = "2026-10-15T12:02:00Z".into();
        let mut other_account = event.clone();
        other_account.account_id = uuid::Uuid::new_v4().to_string();
        assert!(claim(&db, &event).unwrap());
        assert!(claim(&db, &later).unwrap());
        assert!(claim(&db, &other_account).unwrap());
        release_claim(&db, &event).unwrap();
        assert!(claim(&db, &event).unwrap());
        assert!(!claim(&db, &later).unwrap());
        assert!(!claim(&db, &other_account).unwrap());
    }
    #[test]
    fn activation_uses_only_the_current_validated_origin_and_identifiers() {
        let event = notification();
        let url = destination(
            Url::parse("https://engineering.fortyone.app/my-work?old=true#fragment").unwrap(),
            &event,
            true,
        );
        assert_eq!(url.host_str(), Some("engineering.fortyone.app"));
        assert_eq!(
            url.path(),
            format!("/notifications/{}", event.notification_id)
        );
        assert!(url.fragment().is_none());
        assert!(!url.query().unwrap().contains("old="));
        let local = destination(
            Url::parse("http://localhost:3000/engineering/my-work").unwrap(),
            &event,
            false,
        );
        assert_eq!(
            local.path(),
            format!("/engineering/notifications/{}", event.notification_id)
        );
    }
    #[test]
    fn unsupported_or_oversized_content_is_rejected() {
        let mut event = notification();
        event.entity_type = "https://outside.example".into();
        assert!(event.validate().is_err());
        event.entity_type = "story".into();
        event.body = "x".repeat(8193);
        assert!(event.validate().is_err());
    }

    #[test]
    fn comment_activation_preserves_the_comment_across_system_delivery() {
        let mut event = notification();
        let comment_id = uuid::Uuid::new_v4().to_string();
        event.comment_id = Some(comment_id.clone());
        assert!(event.validate().is_ok());
        let restored: Notification = serde_json::from_value(activation_payload(&event)).unwrap();
        assert_eq!(restored.comment_id.as_deref(), Some(comment_id.as_str()));
        for (origin, production) in [
            ("https://engineering.fortyone.app/my-work", true),
            ("http://localhost:3000/engineering/my-work", false),
        ] {
            let url = destination(Url::parse(origin).unwrap(), &restored, production);
            assert!(url
                .query_pairs()
                .any(|(key, value)| key == "comment" && value == comment_id));
        }
    }

    #[test]
    fn older_notification_payloads_remain_valid_without_comment_metadata() {
        let event = notification();
        let serialized = serde_json::to_value(&event).unwrap();
        assert!(serialized.get("commentId").is_none());
        let restored: Notification = serde_json::from_value(serialized).unwrap();
        assert!(restored.validate().is_ok());
        assert!(restored.comment_id.is_none());
    }

    #[test]
    fn invalid_comment_targets_are_rejected() {
        for value in ["not-a-uuid", "00000000-0000-0000-0000-000000000000"] {
            let mut event = notification();
            event.comment_id = Some(value.into());
            assert!(event.validate().is_err());
        }
        let mut event = notification();
        event.entity_type = "objective".into();
        event.comment_id = Some(uuid::Uuid::new_v4().to_string());
        assert!(event.validate().is_err());
    }
}
