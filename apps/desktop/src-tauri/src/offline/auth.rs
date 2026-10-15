use super::{model::now, runtime, transport::decode, Runtime};
use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine};
use rand::RngCore;
use serde_json::json;
use sha2::{Digest, Sha256};
use std::sync::Arc;
use tauri::{AppHandle, Emitter, Manager, Url, WebviewWindow};
use tauri_plugin_opener::OpenerExt;

pub const CALLBACK_SCHEME: &str = if cfg!(feature = "production") {
    "fortyone-desktop"
} else {
    "fortyone-desktop-dev"
};

fn redirect_uri() -> String {
    format!("{CALLBACK_SCHEME}://login")
}

pub struct Login {
    state: String,
    verifier: String,
    started: u64,
    account: Option<String>,
}

fn random_value() -> String {
    let mut bytes = [0u8; 32];
    rand::rngs::OsRng.fill_bytes(&mut bytes);
    URL_SAFE_NO_PAD.encode(bytes)
}

#[tauri::command]
pub fn desktop_sign_in(window: WebviewWindow) -> Result<(), String> {
    let runtime = runtime(&window)?;
    let login = Login {
        state: random_value(),
        verifier: random_value(),
        started: now(),
        account: runtime.store.account().ok(),
    };
    let challenge = URL_SAFE_NO_PAD.encode(Sha256::digest(login.verifier.as_bytes()));
    let mut url = Url::parse(&format!("{}/auth/desktop", runtime.transport.website))
        .map_err(|e| e.to_string())?;
    url.query_pairs_mut()
        .append_pair("state", &login.state)
        .append_pair("code_challenge", &challenge)
        .append_pair("redirect_uri", &redirect_uri());
    *runtime.login.lock().map_err(|e| e.to_string())? = Some(login);
    window
        .app_handle()
        .opener()
        .open_url(url.as_str(), None::<&str>)
        .map_err(|_| "Could not open your browser".into())
}

pub async fn callback(app: AppHandle, url: Url) -> Result<(), String> {
    if url.scheme() != CALLBACK_SCHEME
        || url.host_str() != Some("login")
        || !matches!(url.path(), "" | "/")
        || url.fragment().is_some()
        || !url.username().is_empty()
        || url.password().is_some()
    {
        return Err("Invalid desktop sign-in link".into());
    }
    let params: Vec<_> = url.query_pairs().collect();
    if params.len() != 2 {
        return Err("Invalid sign-in response".into());
    }
    let code = params
        .iter()
        .find(|(k, _)| k == "code")
        .map(|(_, v)| v.as_ref())
        .ok_or("Missing authorization code")?;
    let state = params
        .iter()
        .find(|(k, _)| k == "state")
        .map(|(_, v)| v.as_ref())
        .ok_or("Missing authorization state")?;
    let runtime = app.state::<Arc<Runtime>>().inner().clone();
    let _guard = runtime.sync.lock().await;
    let login = {
        let mut pending = runtime.login.lock().map_err(|e| e.to_string())?;
        let login = pending
            .as_ref()
            .ok_or("Start sign-in from the desktop app again")?;
        if login.state != state || now().saturating_sub(login.started) > 300 {
            return Err("Sign-in expired or belongs to another request".into());
        }
        pending.take().ok_or("Sign-in already completed")?
    };
    if runtime.store.account().ok() != login.account {
        return Err("Your account changed. Start sign-in again.".into());
    }
    let response=runtime.transport.client.post(format!("{}/auth/desktop/exchange",runtime.transport.api)).header(reqwest::header::ORIGIN,&runtime.transport.website).json(&json!({"code":code,"state":state,"codeVerifier":login.verifier,"redirectUri":redirect_uri()})).send().await.map_err(|_|"Could not complete sign-in. Start again when connected.".to_owned())?;
    let cookie = response
        .cookies()
        .find(|c| c.name() == "fortyone_session")
        .map(|c| c.value().to_owned());
    let user = decode(response).await?;
    let account = user["id"].as_str().ok_or("Invalid account response")?;
    if login
        .account
        .as_deref()
        .is_some_and(|previous| previous != account)
    {
        return Err("Sign out of the desktop app before signing in with another account. Your local work is still saved.".into());
    }
    let token = cookie.ok_or("The server did not establish a desktop session")?;
    let workspaces = runtime.transport.get("/workspaces", &token).await?;
    let previous = if runtime.store.account().ok().as_deref() == Some(account) {
        runtime.store.setting("workspace")?
    } else {
        None
    };
    let memberships = workspaces.as_array().ok_or("Invalid workspace response")?;
    let preferred = previous
        .as_ref()
        .and_then(|v| v.as_str())
        .or_else(|| user["lastUsedWorkspaceId"].as_str());
    let selected = preferred
        .and_then(|id| {
            memberships
                .iter()
                .find(|workspace| workspace["id"].as_str() == Some(id))
        })
        .or_else(|| memberships.first())
        .map(|workspace| workspace["id"].clone())
        .unwrap_or(serde_json::Value::Null);
    runtime.transport.replace_session(account, token).await?;
    runtime.store.save_account(&user, &workspaces)?;
    runtime.store.set_setting("workspace", &selected)?;
    app.emit("desktop-session-restored", json!({"accountId": account}))
        .map_err(|e| e.to_string())?;
    app.emit("desktop-state-changed", ())
        .map_err(|e| e.to_string())?;
    if let Some(window) = app.get_webview_window("main") {
        let _ = window.show();
        let _ = window.set_focus();
    }
    Ok(())
}
