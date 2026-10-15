use serde_json::Value;
use tauri::{Url, WebviewWindow};

use super::{runtime, transport::decode};

fn collaboration_url(value: &str) -> Result<Url, String> {
    let url = Url::parse(value).map_err(|_| "Invalid collaboration service address")?;
    let development = !cfg!(feature = "production")
        && url.scheme() == "ws"
        && matches!(url.host_str(), Some("localhost" | "127.0.0.1"));
    if (url.scheme() != "wss" && !development)
        || url.host_str().is_none()
        || !url.username().is_empty()
        || url.password().is_some()
        || url.query().is_some()
        || url.fragment().is_some()
    {
        return Err("The collaboration service must use a trusted secure address".into());
    }
    Ok(url)
}

#[tauri::command]
pub async fn desktop_collaboration_session(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    document_id: String,
) -> Result<Value, String> {
    let runtime = runtime(&window)?;
    let account = super::require_account(&runtime, &expected_account_id)?;
    let workspace = runtime.store.authorize(&account, &workspace_id)?;
    uuid::Uuid::parse_str(&document_id).map_err(|_| "Invalid document")?;
    let slug = workspace["slug"]
        .as_str()
        .filter(|slug| {
            !slug.is_empty() && slug.bytes().all(|b| b.is_ascii_alphanumeric() || b == b'-')
        })
        .ok_or("Invalid workspace address")?;
    let configured = option_env!("FORTYONE_COLLABORATION_URL")
        .filter(|v| !v.is_empty())
        .ok_or(
            "Live document editing is not configured for this build. Your draft is saved locally.",
        )?;
    let url = collaboration_url(configured)?;
    let token = runtime.transport.token(&account).await?;
    let response = runtime
        .transport
        .request(
            reqwest::Method::POST,
            &format!("/workspaces/{slug}/documents/{document_id}/collaboration-session"),
            &token,
        )?
        .send()
        .await
        .map_err(|_| "Could not connect to document collaboration")?;
    let mut ticket = decode(response).await?;
    if runtime.store.account()? != account {
        return Err("Your account changed. Reopen this document.".into());
    }
    if ticket["token"].as_str().is_none() || ticket["name"].as_str().is_none() {
        return Err("Invalid document session response".into());
    }
    // Only a short-lived, document-scoped ticket reaches the webview. The
    // account session remains in the OS credential store and Rust transport.
    ticket["url"] = Value::String(url.to_string());
    Ok(ticket)
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn collaboration_addresses_cannot_embed_secrets_or_use_plain_remote_transport() {
        assert!(collaboration_url("wss://collaboration.example/socket").is_ok());
        for value in [
            "https://example.com",
            "ws://example.com",
            "wss://user:pass@example.com",
            "wss://example.com?token=secret",
        ] {
            assert!(collaboration_url(value).is_err());
        }
    }
}
