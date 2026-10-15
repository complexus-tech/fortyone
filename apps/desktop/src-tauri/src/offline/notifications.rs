//! Native-owned subscriptions to the existing authenticated workspace SSE API.
//! The streams survive hidden windows; no process means no notification delivery.
mod cache;
mod codec;
#[cfg(test)]
mod tests;

use super::{model::Settings, Runtime};
use cache::{badge, catch_up, initialize, prune_receipts, update_cache};
use codec::{activation_target, decode_event, Notification, RealtimeEvent, SseDecoder};
use reqwest::{header, Method};
use serde_json::{json, Value};
use std::{
    collections::HashMap,
    sync::{Arc, Mutex},
    time::{Duration, Instant},
};
use tauri::{AppHandle, Emitter, Listener, Manager};
use tauri_plugin_notification::{NotificationExt, PermissionState};

fn current(runtime: &Runtime, account: &str, workspace: &str) -> bool {
    runtime.store.account().ok().as_deref() == Some(account)
        && runtime.store.authorize(account, workspace).is_ok()
}

fn preferences_allow(runtime: &Runtime) -> bool {
    runtime
        .store
        .setting("preferences")
        .ok()
        .flatten()
        .and_then(|v| serde_json::from_value::<Settings>(v).ok())
        .unwrap_or_default()
        .notifications
}

fn clean_text(value: &str, limit: usize) -> String {
    value
        .chars()
        .filter(|character| !character.is_control())
        .take(limit)
        .collect()
}

fn deliver(app: &AppHandle, runtime: &Runtime, notification: &Notification) {
    if notification.read_at.is_some()
        || !preferences_allow(runtime)
        || !current(
            runtime,
            &notification.recipient_id,
            &notification.workspace_id,
        )
    {
        return;
    }
    if app.notification().permission_state().ok() != Some(PermissionState::Granted) {
        return;
    }
    let result = app
        .notification()
        .builder()
        .id((uuid::Uuid::parse_str(&notification.id)
            .map(|v| v.as_u128() as u32)
            .unwrap_or_default()
            & 0x7fff_ffff) as i32)
        .title("FortyOne")
        .body(clean_text(&notification.title, 240))
        .extra("fortyoneTarget", notification.activation())
        .show();
    if result.is_err() {
        let _ = app.emit(
            "desktop-notification-error",
            "The system could not display a notification. Your inbox is up to date.",
        );
    }
}

async fn stream(
    app: &AppHandle,
    runtime: &Runtime,
    account: &str,
    workspace: &str,
    slug: &str,
) -> Result<(), String> {
    let token = runtime
        .transport
        .token(account)
        .await
        .map_err(|_| "access-required")?;
    let mut response = runtime
        .transport
        .request(
            Method::GET,
            &format!("/workspaces/{slug}/notifications/subscribe"),
            &token,
        )?
        .header(header::ACCEPT, "text/event-stream")
        // Override the ordinary API's 30s total timeout. The 60s chunk deadline
        // detects a lost connection; the server sends a keep-alive every 25s.
        .timeout(Duration::from_secs(24 * 3600))
        .send()
        .await
        .map_err(|_| "The realtime connection was interrupted")?;
    if response.status().as_u16() == 403 {
        runtime.store.deny_workspace(account, workspace)?;
        badge(app, runtime);
    }
    if matches!(response.status().as_u16(), 401 | 403) {
        return Err("access-required".into());
    }
    if !response.status().is_success()
        || !response
            .headers()
            .get(header::CONTENT_TYPE)
            .and_then(|v| v.to_str().ok())
            .is_some_and(|v| v.starts_with("text/event-stream"))
    {
        return Err("The realtime service is not ready".into());
    }
    catch_up(app, runtime, account, workspace, slug, &token).await?;
    let _ = app.emit(
        "desktop-realtime-status",
        json!({"workspaceId":workspace,"status":"connected"}),
    );
    let mut decoder = SseDecoder::default();
    loop {
        let chunk = tokio::time::timeout(Duration::from_secs(60), response.chunk())
            .await
            .map_err(|_| "The realtime connection timed out")?
            .map_err(|_| "The realtime connection was interrupted")?;
        let Some(chunk) = chunk else {
            return Err("The realtime connection closed".into());
        };
        if !current(runtime, account, workspace) {
            return Err("access-required".into());
        }
        for serialized in decoder.push(&chunk)? {
            let event = decode_event(&serialized, account, workspace)?;
            let fresh = update_cache(&runtime.store, account, workspace, &event)?;
            if fresh {
                if let RealtimeEvent::Notification(notification) = &event {
                    deliver(app, runtime, notification);
                }
            }
            badge(app, runtime);
            let payload: Value =
                serde_json::from_str(&serialized).map_err(|_| "Invalid realtime event")?;
            let _ = app.emit(
                "desktop-realtime",
                json!({"workspaceId":workspace,"workspaceSlug":slug,"event":payload}),
            );
        }
    }
}

async fn subscribe(
    app: AppHandle,
    runtime: Arc<Runtime>,
    account: String,
    workspace: String,
    slug: String,
) {
    let mut backoff = 1u64;
    loop {
        if !current(&runtime, &account, &workspace) {
            return;
        }
        let started = Instant::now();
        let result = stream(&app, &runtime, &account, &workspace, &slug).await;
        if result
            .as_ref()
            .err()
            .is_some_and(|error| error == "access-required")
        {
            let _ = app.emit(
                "desktop-realtime-status",
                json!({"workspaceId":workspace,"status":"access-required"}),
            );
            return;
        }
        let _ = app.emit(
            "desktop-realtime-status",
            json!({"workspaceId":workspace,"status":"reconnecting"}),
        );
        if started.elapsed() > Duration::from_secs(60) {
            backoff = 1;
        }
        let jitter = u64::from(rand::random::<u16>()) % 1000;
        tokio::time::sleep(Duration::from_millis(backoff * 1000 + jitter)).await;
        backoff = (backoff * 2).min(60);
    }
}

type Subscriptions = Mutex<HashMap<(String, String, String), tauri::async_runtime::JoinHandle<()>>>;
fn reconcile(app: &AppHandle, runtime: &Arc<Runtime>, subscriptions: &Subscriptions) {
    let bootstrap = runtime.store.bootstrap().unwrap_or(Value::Null);
    let account = runtime.store.account().ok();
    let mut desired = Vec::new();
    if let Some(account) = account {
        if let Err(error) = prune_receipts(&runtime.store, &account) {
            let _ = app.emit("desktop-error", error);
        }
        for workspace in bootstrap["workspaces"].as_array().into_iter().flatten() {
            if let (Some(id), Some(slug)) = (workspace["id"].as_str(), workspace["slug"].as_str()) {
                if uuid::Uuid::parse_str(id).is_ok()
                    && !slug.is_empty()
                    && slug.len() <= 120
                    && slug.bytes().all(|b| b.is_ascii_alphanumeric() || b == b'-')
                    && runtime.store.authorize(&account, id).is_ok()
                {
                    desired.push((account.clone(), id.to_owned(), slug.to_owned()));
                }
            }
        }
    }
    let Ok(mut running) = subscriptions.lock() else {
        return;
    };
    running.retain(|key, handle| {
        if desired.contains(key) && !handle.inner().is_finished() {
            return true;
        }
        handle.abort();
        false
    });
    for key in desired {
        if !running.contains_key(&key) {
            let (account, workspace, slug) = key.clone();
            running.insert(
                key,
                tauri::async_runtime::spawn(subscribe(
                    app.clone(),
                    runtime.clone(),
                    account,
                    workspace,
                    slug,
                )),
            );
        }
    }
    drop(running);
    badge(app, runtime);
}

pub fn start(
    app: &AppHandle,
    runtime: Arc<Runtime>,
    activate: fn(&AppHandle, Value) -> Result<(), String>,
) -> Result<(), String> {
    initialize(&runtime.store)?;
    let subscriptions = Arc::new(Subscriptions::default());
    let (observer_app, observer_runtime, observer_subscriptions) =
        (app.clone(), runtime.clone(), subscriptions.clone());
    app.listen("desktop-state-changed", move |_| {
        reconcile(&observer_app, &observer_runtime, &observer_subscriptions)
    });
    let (activation_app, activation_runtime) = (app.clone(), runtime.clone());
    app.notification()
        .on_action(move |action| {
            if action.action_id() != "tap" {
                return;
            }
            let Some(value) = action
                .notification()
                .and_then(|notification| notification.extra().get("fortyoneTarget"))
            else {
                return;
            };
            let Ok(account) = activation_runtime.store.account() else {
                return;
            };
            let Some(target) = activation_target(value, &account) else {
                return;
            };
            let Some(workspace) = target["workspaceId"].as_str() else {
                return;
            };
            if !current(&activation_runtime, &account, workspace) {
                return;
            }
            if let Err(error) = activate(&activation_app, target) {
                let _ = activation_app.emit("desktop-error", error);
            }
        })
        .map_err(|e| e.to_string())?;
    reconcile(app, &runtime, &subscriptions);
    Ok(())
}

#[tauri::command]
pub async fn desktop_mark_notification_read(
    window: tauri::WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    notification_id: String,
    observed_created_at: String,
) -> Result<Value, String> {
    let runtime = super::runtime(&window)?;
    let account = super::require_account(&runtime, &expected_account_id)?;
    let workspace = runtime.store.authorize(&account, &workspace_id)?;
    uuid::Uuid::parse_str(&notification_id).map_err(|_| "Invalid notification")?;
    if observed_created_at.is_empty()
        || observed_created_at.len() > 64
        || observed_created_at.chars().any(char::is_control)
    {
        return Err("Refresh the notification before marking it as read".into());
    }
    let slug = workspace["slug"]
        .as_str()
        .filter(|value| {
            !value.is_empty()
                && value.len() <= 120
                && value
                    .bytes()
                    .all(|byte| byte.is_ascii_alphanumeric() || byte == b'-')
        })
        .ok_or("Invalid workspace address")?;
    let token = runtime.transport.token(&account).await?;
    let response = runtime
        .transport
        .request(
            Method::PUT,
            &format!("/workspaces/{slug}/notifications/{notification_id}/read"),
            &token,
        )?
        .query(&[("observedCreatedAt", &observed_created_at)])
        .send()
        .await
        .map_err(|_| "Connect to the internet to mark this notification as read")?;
    super::transport::decode(response).await?;
    if !current(&runtime, &account, &workspace_id) {
        return Err("Account or workspace access changed".into());
    }
    catch_up(
        window.app_handle(),
        &runtime,
        &account,
        &workspace_id,
        slug,
        &token,
    )
    .await?;
    runtime.store.workspace(&account, &workspace_id)
}
