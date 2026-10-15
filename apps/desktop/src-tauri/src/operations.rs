//! Shared durable operation contract and dependency ordering.
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct Operation {
    pub id: String,
    pub kind: String,
    pub resource_id: String,
    pub payload: Value,
    pub base_version: Option<String>,
    pub status: String,
    pub error: Option<String>,
    #[serde(default)]
    pub server_version: Option<String>,
    #[serde(default)]
    pub server_value: Option<Value>,
    #[serde(default)]
    pub attempted: bool,
}

pub fn next_operation(operations: Vec<Operation>) -> Option<Operation> {
    let mut blocked = std::collections::HashSet::new();
    for operation in operations {
        let dependency_blocked = blocked.contains(&operation.resource_id)
            || ["storyId", "documentId", "parentId"].iter().any(|key| {
                operation.payload[*key]
                    .as_str()
                    .is_some_and(|id| blocked.contains(id))
            });
        if operation.status == "pending" && !dependency_blocked {
            return Some(operation);
        }
        blocked.insert(operation.resource_id.clone());
    }
    None
}

pub fn validate_operation(operation: &Operation) -> Result<(), String> {
    uuid::Uuid::parse_str(&operation.resource_id).map_err(|_| "Invalid resource identifier")?;
    if !matches!(
        operation.kind.as_str(),
        "story.create"
            | "story.update"
            | "comment.create"
            | "document.create"
            | "document.update"
            | "attachment.upload"
    ) {
        return Err("This change is not supported offline".into());
    }
    if !operation.payload.is_object() || operation.payload.to_string().len() > 2_000_000 {
        return Err("The change is invalid or exceeds the local save limit".into());
    }
    if operation.kind.ends_with(".update")
        && operation.base_version.as_deref().is_none_or(str::is_empty)
    {
        return Err("Refresh this item before editing it".into());
    }
    Ok(())
}
