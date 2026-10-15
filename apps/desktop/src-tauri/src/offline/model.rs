use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::time::{SystemTime, UNIX_EPOCH};

pub fn now() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs()
}

pub use crate::operations::{next_operation, validate_operation, Operation};

pub fn empty_workspace(workspace_id: &str) -> Value {
    json!({"workspaceId":workspace_id,"stories":[],"documents":[],"comments":[],"teams":[],"statuses":[],"pending":[],"lastSyncedAt":null})
}

pub fn apply_operation(snapshot: &mut Value, operation: &Operation) {
    let collection = match operation.kind.as_str() {
        "story.create" | "story.update" => "stories",
        "document.create" | "document.update" => "documents",
        "comment.create" => "comments",
        _ => return,
    };
    let Some(items) = snapshot.get_mut(collection).and_then(Value::as_array_mut) else {
        return;
    };
    if let Some(item) = items
        .iter_mut()
        .find(|item| item["id"].as_str() == Some(&operation.resource_id))
    {
        if let (Some(target), Some(patch)) = (item.as_object_mut(), operation.payload.as_object()) {
            target.extend(patch.clone());
            target.insert("pendingSync".into(), Value::Bool(true));
        }
    } else if operation.kind.ends_with(".create") {
        let mut item = operation.payload.clone();
        item["id"] = json!(operation.resource_id);
        item["pendingSync"] = json!(true);
        items.push(item);
    }
}

#[derive(Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Settings {
    pub notifications: bool,
    pub launch_at_login: bool,
    pub theme: String,
    #[serde(default)]
    pub app_version: String,
    #[serde(default)]
    pub update_configured: bool,
}

impl Default for Settings {
    fn default() -> Self {
        Self {
            notifications: false,
            launch_at_login: false,
            theme: "system".into(),
            app_version: env!("CARGO_PKG_VERSION").into(),
            update_configured: false,
        }
    }
}
