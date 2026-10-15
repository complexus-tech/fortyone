use rusqlite::{params, OptionalExtension};
use serde_json::{json, Value};
use tauri::{Url, WebviewWindow};
use tauri_plugin_opener::OpenerExt;

use super::{runtime, Runtime};

pub fn workspace_slug(workspace: &Value) -> Result<&str, String> {
    workspace["slug"]
        .as_str()
        .filter(|value| {
            !value.is_empty()
                && value
                    .bytes()
                    .all(|b| b.is_ascii_alphanumeric() || b == b'-')
        })
        .ok_or_else(|| "Invalid workspace address".into())
}

pub fn read(
    runtime: &Runtime,
    workspace_id: &str,
    expected_account_id: &str,
) -> Result<Value, String> {
    let account = super::require_account(runtime, expected_account_id)?;
    let workspace = runtime.store.authorize(&account, workspace_id)?;
    let cached: Option<String> = runtime
        .store
        .0
        .lock()
        .map_err(|e| e.to_string())?
        .query_row(
            "SELECT data FROM workspace_settings WHERE account=? AND workspace=?",
            params![account, workspace_id],
            |r| r.get(0),
        )
        .optional()
        .map_err(|e| e.to_string())?;
    if let Some(cached) = cached {
        return serde_json::from_str(&cached).map_err(|e| e.to_string());
    }
    let profile = {
        let db = runtime.store.0.lock().map_err(|error| error.to_string())?;
        super::require_account_in(&db, &account)?;
        let encoded: String = db
            .query_row("SELECT user FROM accounts WHERE id=?", [&account], |row| {
                row.get(0)
            })
            .map_err(|error| error.to_string())?;
        serde_json::from_str::<Value>(&encoded).map_err(|error| error.to_string())?
    };
    let snapshot = runtime.store.workspace(&account, workspace_id)?;
    Ok(
        json!({"profile":profile,"workspace":workspace,"teams":snapshot["teams"],"members":[],"preferences":{},"downloadedAt":null}),
    )
}

pub async fn synchronize(runtime: &Runtime, workspace_id: &str) -> Result<Value, String> {
    let account = runtime.store.account()?;
    synchronize_for(runtime, workspace_id, &account).await
}

async fn synchronize_for(
    runtime: &Runtime,
    workspace_id: &str,
    expected_account_id: &str,
) -> Result<Value, String> {
    let _guard = runtime.sync.lock().await;
    let account = super::require_account(runtime, expected_account_id)?;
    let workspace = runtime.store.authorize(&account, workspace_id)?;
    let slug = workspace_slug(&workspace)?;
    let token = runtime.transport.token(&account).await?;
    let response = runtime
        .transport
        .request(
            reqwest::Method::GET,
            &format!("/workspaces/{slug}/desktop/settings"),
            &token,
        )?
        .send()
        .await
        .map_err(|_| {
            "Could not download workspace settings. Saved settings remain available.".to_owned()
        })?;
    if matches!(response.status().as_u16(), 403 | 404) {
        runtime.store.deny_workspace(&account, workspace_id)?;
    }
    let data = super::transport::decode(response).await?;
    if runtime.store.account()? != account {
        return Err("Your account changed while settings were downloading".into());
    }
    if data["profile"]["id"].as_str() != Some(&account)
        || data["workspace"]["id"].as_str() != Some(workspace_id)
    {
        return Err("The server returned settings for another account or workspace".into());
    }
    if !data["profile"].is_object()
        || !data["workspace"].is_object()
        || !data["teams"].is_array()
        || !data["members"].is_array()
        || !data["preferences"].is_object()
    {
        return Err("Incomplete settings response".into());
    }
    runtime.store.0.lock().map_err(|e|e.to_string())?.execute("INSERT INTO workspace_settings VALUES(?,?,?) ON CONFLICT(account,workspace) DO UPDATE SET data=excluded.data",params![account,workspace_id,data.to_string()]).map_err(|e|e.to_string())?;
    Ok(data)
}

#[tauri::command]
pub fn desktop_read_settings(
    window: WebviewWindow,
    workspace_id: String,
    expected_account_id: String,
) -> Result<Value, String> {
    read(
        runtime(&window)?.as_ref(),
        &workspace_id,
        &expected_account_id,
    )
}

#[tauri::command]
pub async fn desktop_sync_settings(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
) -> Result<Value, String> {
    synchronize_for(
        runtime(&window)?.as_ref(),
        &workspace_id,
        &expected_account_id,
    )
    .await
}

pub fn workspace_url(workspace: &Value, path: &str, production: bool) -> Result<Url, String> {
    if !path.starts_with('/')
        || path.starts_with("//")
        || path.len() > 1024
        || path.split('/').any(|part| {
            part == "."
                || part == ".."
                || !part
                    .bytes()
                    .all(|c| c.is_ascii_alphanumeric() || matches!(c, b'-' | b'_'))
        })
    {
        return Err("Invalid workspace page".into());
    }
    let slug = workspace_slug(workspace)?;
    let destination = if production {
        format!("https://{slug}.fortyone.app{path}")
    } else {
        format!("http://localhost:3000/{slug}{path}")
    };
    Url::parse(&destination).map_err(|_| "Invalid workspace page".into())
}

#[tauri::command]
pub fn desktop_open_workspace_page(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    path: String,
) -> Result<(), String> {
    let runtime = runtime(&window)?;
    let account = super::require_account(&runtime, &expected_account_id)?;
    let workspace = runtime.store.authorize(&account, &workspace_id)?;
    let url = workspace_url(&workspace, &path, cfg!(feature = "production"))?;
    window
        .opener()
        .open_url(url.as_str(), None::<&str>)
        .map_err(|e| e.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn workspace_links_cannot_escape_the_authorized_origin() {
        let workspace = json!({"slug":"acme"});
        assert_eq!(
            workspace_url(&workspace, "/settings/account", true)
                .unwrap()
                .as_str(),
            "https://acme.fortyone.app/settings/account"
        );
        for path in [
            "//evil.example",
            "/../admin",
            "/%2e%2e/admin",
            "/settings?token=secret",
            "/settings#fragment",
            "https://evil.example",
        ] {
            assert!(workspace_url(&workspace, path, true).is_err());
        }
    }
}
