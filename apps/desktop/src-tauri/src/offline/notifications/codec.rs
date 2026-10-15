use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

pub(super) const MAX_EVENT_BYTES: usize = 128 * 1024;

#[derive(Default)]
pub(super) struct SseDecoder {
    line: Vec<u8>,
    data: String,
    event: String,
}
impl SseDecoder {
    pub(super) fn push(&mut self, bytes: &[u8]) -> Result<Vec<String>, String> {
        let mut events = Vec::new();
        for byte in bytes {
            if *byte != b'\n' {
                self.line.push(*byte);
                if self.line.len() + self.data.len() > MAX_EVENT_BYTES {
                    return Err("The event stream exceeded its size limit".into());
                }
                continue;
            }
            let line = std::str::from_utf8(&self.line)
                .map_err(|_| "Invalid event encoding")?
                .trim_end_matches('\r');
            if line.is_empty() {
                if !self.data.is_empty() && (self.event.is_empty() || self.event == "message") {
                    events.push(std::mem::take(&mut self.data));
                }
                self.data.clear();
                self.event.clear();
            } else if let Some(value) = line.strip_prefix("data:") {
                if !self.data.is_empty() {
                    self.data.push('\n');
                }
                self.data.push_str(value.strip_prefix(' ').unwrap_or(value));
            } else if let Some(value) = line.strip_prefix("event:") {
                self.event = value.trim_start().to_owned();
            }
            self.line.clear();
        }
        Ok(events)
    }
}

#[derive(Clone, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub(super) struct Notification {
    pub(super) id: String,
    #[serde(alias = "recipient_id")]
    pub(super) recipient_id: String,
    #[serde(alias = "workspace_id")]
    pub(super) workspace_id: String,
    #[serde(alias = "entity_type")]
    pub(super) entity_type: String,
    #[serde(alias = "entity_id")]
    pub(super) entity_id: String,
    pub(super) title: String,
    #[serde(default)]
    pub(super) message: Value,
    #[serde(alias = "created_at")]
    pub(super) created_at: String,
    #[serde(default, alias = "read_at")]
    pub(super) read_at: Option<String>,
    #[serde(rename = "type")]
    pub(super) notification_type: String,
}
impl Notification {
    pub(super) fn valid_for(&self, account: &str, workspace: &str) -> bool {
        self.recipient_id == account
            && self.workspace_id == workspace
            && [
                &self.id,
                &self.recipient_id,
                &self.workspace_id,
                &self.entity_id,
            ]
            .iter()
            .all(|id| uuid::Uuid::parse_str(id).is_ok())
            && matches!(
                self.entity_type.as_str(),
                "story" | "comment" | "objective" | "key_result"
            )
            && matches!(
                self.notification_type.as_str(),
                "story_update"
                    | "story_comment"
                    | "comment_reply"
                    | "objective_update"
                    | "key_result_update"
                    | "mention"
            )
            && !self.title.trim().is_empty()
            && self.title.len() <= 2048
            && !self.created_at.is_empty()
            && self.created_at.len() <= 64
    }

    pub(super) fn activation(&self) -> Value {
        // Unsupported resources open the workspace inbox. Never follow server-provided URLs.
        json!({"accountId":self.recipient_id,"workspaceId":self.workspace_id,
            "notificationId":self.id,"kind":if self.entity_type == "story" { "story" } else { "notifications" },
            "resourceId":if self.entity_type == "story" { Some(&self.entity_id) } else { None }})
    }
}

#[derive(Deserialize)]
pub(super) struct StoryUpdate {
    #[serde(rename = "storyId")]
    pub(super) story_id: String,
    pub(super) changes: serde_json::Map<String, Value>,
}

pub(super) enum RealtimeEvent {
    Notification(Notification),
    Story(StoryUpdate),
    Calendar,
}
pub(super) fn decode_event(
    serialized: &str,
    account: &str,
    workspace: &str,
) -> Result<RealtimeEvent, String> {
    let value: Value = serde_json::from_str(serialized).map_err(|_| "Invalid realtime event")?;
    match value["type"].as_str() {
        Some("calendar.updated") => Ok(RealtimeEvent::Calendar),
        Some("story.workspace_update") => {
            let update: StoryUpdate =
                serde_json::from_value(value).map_err(|_| "Invalid story event")?;
            if uuid::Uuid::parse_str(&update.story_id).is_err()
                || update.changes.is_empty()
                || !update
                    .changes
                    .iter()
                    .all(|(key, value)| valid_story_change(key, value))
            {
                return Err("Invalid story update fields".into());
            }
            Ok(RealtimeEvent::Story(update))
        }
        _ => {
            let notification: Notification =
                serde_json::from_value(value).map_err(|_| "Invalid notification event")?;
            if !notification.valid_for(account, workspace) {
                return Err("Notification scope did not match the subscription".into());
            }
            Ok(RealtimeEvent::Notification(notification))
        }
    }
}
fn valid_story_change(key: &str, value: &Value) -> bool {
    match key {
        "checklistChanged" => value == &Value::Bool(true),
        "autoSchedulingEnabled" | "autoSchedulingLocked" => value.is_boolean(),
        "assigneeId" => {
            value.is_null()
                || value
                    .as_str()
                    .is_some_and(|v| uuid::Uuid::parse_str(v).is_ok())
        }
        "statusId" => value
            .as_str()
            .is_some_and(|v| uuid::Uuid::parse_str(v).is_ok()),
        "title" => value
            .as_str()
            .is_some_and(|v| !v.trim().is_empty() && v.len() <= 8192),
        "priority" => matches!(
            value.as_str(),
            Some("No Priority" | "Urgent" | "High" | "Medium" | "Low")
        ),
        "autoSchedulingStatus" => matches!(
            value.as_str(),
            Some(
                "off"
                    | "needs_owner"
                    | "needs_time"
                    | "planning"
                    | "scheduled"
                    | "at_risk"
                    | "cannot_fit"
                    | "locked"
            )
        ),
        "completedAt" | "autoSchedulingReason" | "autoSchedulingUpdatedAt" => {
            value.is_null() || value.as_str().is_some_and(|v| v.len() <= 8192)
        }
        _ => false,
    }
}

pub(super) fn activation_target(value: &Value, account: &str) -> Option<Value> {
    if value["accountId"].as_str() != Some(account) {
        return None;
    }
    for key in ["workspaceId", "notificationId"] {
        uuid::Uuid::parse_str(value[key].as_str()?).ok()?;
    }
    let kind = value["kind"].as_str()?;
    if kind == "story" {
        uuid::Uuid::parse_str(value["resourceId"].as_str()?).ok()?;
    } else if kind != "notifications" {
        return None;
    }
    Some(
        json!({"accountId":account,"workspaceId":value["workspaceId"],"notificationId":value["notificationId"],"kind":kind,"resourceId":if kind == "story" {value["resourceId"].clone()} else {Value::Null}}),
    )
}
