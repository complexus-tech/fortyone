use super::{model::now, settings, transport::decode, Runtime};
use serde_json::{json, Value};
use std::{collections::HashSet, time::Duration};
use tauri::{AppHandle, Emitter};

pub fn state_key(account: &str, workspace: &str) -> String {
    format!("download:{account}:{workspace}")
}

pub(super) fn progress(
    app: &AppHandle,
    runtime: &Runtime,
    account: &str,
    workspace: &str,
    stage: &str,
    count: usize,
    error: Option<&str>,
) -> Result<(), String> {
    let value = json!({"workspaceId":workspace,"stage":stage,"downloaded":count,"updatedAt":now(),"error":error});
    runtime
        .store
        .set_setting(&state_key(account, workspace), &value)?;
    app.emit("desktop-download-progress", value)
        .map_err(|e| e.to_string())
}

pub async fn fetch(
    runtime: &Runtime,
    account: &str,
    workspace: &str,
    query: &[(&str, String)],
) -> Result<Value, String> {
    let descriptor = runtime.store.membership(account, workspace)?;
    let slug = settings::workspace_slug(&descriptor)?;
    let token = runtime.transport.token(account).await?;
    let response = runtime
        .transport
        .request(
            reqwest::Method::GET,
            &format!("/workspaces/{slug}/desktop/sync"),
            &token,
        )?
        .query(query)
        .send()
        .await
        .map_err(|_| {
            "Could not download the workspace. Saved work remains available.".to_owned()
        })?;
    if matches!(response.status().as_u16(), 403 | 404) {
        runtime.store.deny_workspace(account, workspace)?;
    }
    let data = decode(response).await?;
    if runtime.store.account()? != account {
        return Err("Your account changed while downloading".into());
    }
    if data["workspaceId"].as_str() != Some(workspace) || data["userId"].as_str() != Some(account) {
        return Err("The server returned data for another workspace".into());
    }
    runtime.store.grant_workspace(account, workspace)?;
    runtime.store.snapshot(account, workspace, &data)?;
    Ok(data)
}

pub(super) fn is_current(runtime: &Runtime, account: &str, workspace: &str) -> bool {
    context_matches(&runtime.store, account, workspace)
}

fn context_matches(store: &super::store::Store, account: &str, workspace: &str) -> bool {
    store.account().ok().as_deref() == Some(account)
        && store
            .setting("workspace")
            .ok()
            .flatten()
            .as_ref()
            .and_then(Value::as_str)
            == Some(workspace)
}

fn ensure_current(runtime: &Runtime, account: &str, workspace: &str) -> Result<(), String> {
    if !is_current(runtime, account, workspace) {
        return Err("Your account or workspace changed while downloading".into());
    }
    Ok(())
}

fn pending_changes(runtime: &Runtime, account: &str) -> Result<bool, String> {
    runtime.store.0.lock().map_err(|error|error.to_string())?.query_row(
        "SELECT EXISTS(SELECT 1 FROM operations WHERE account=? AND json_extract(operation,'$.status')='pending')",
        [account],|row|row.get(0)
    ).map_err(|error|error.to_string())
}

async fn complete_threads(
    runtime: &Runtime,
    account: &str,
    workspace: &str,
    snapshot: &Value,
) -> Result<(), String> {
    let complete = snapshot["commentsCompleteFor"]
        .as_array()
        .cloned()
        .unwrap_or_default();
    let descriptor = runtime.store.authorize(account, workspace)?;
    let slug = settings::workspace_slug(&descriptor)?;
    for story in snapshot["stories"].as_array().into_iter().flatten() {
        let id = story["id"].as_str().ok_or("Invalid downloaded story")?;
        if complete.contains(&story["id"]) {
            continue;
        }
        let mut comments = Vec::new();
        let mut bytes = 0;
        for page in 1..=100 {
            let finished = {
                // One request at a time owns the mutation lock. Other edits and
                // synchronization can run between pages of a long discussion.
                let _guard = runtime.sync.lock().await;
                ensure_current(runtime, account, workspace)?;
                runtime.store.authorize(account, workspace)?;
                let token = runtime.transport.token(account).await?;
                let response = runtime
                    .transport
                    .request(
                        reqwest::Method::GET,
                        &format!("/workspaces/{slug}/desktop/stories/{id}/comments"),
                        &token,
                    )?
                    .query(&[("page", page)])
                    .send()
                    .await
                    .map_err(|_| {
                        "Comment download interrupted. Your saved work is retained.".to_owned()
                    })?;
                let data = decode(response).await?;
                ensure_current(runtime, account, workspace)?;
                if data["storyId"].as_str() != Some(id) || data["page"].as_u64() != Some(page) {
                    return Err("Invalid comment download".into());
                }
                let finished = !data["hasMore"]
                    .as_bool()
                    .ok_or("Incomplete comment pagination")?;
                bytes += data.to_string().len();
                if bytes > 8 * 1024 * 1024 {
                    return Err("This discussion exceeds the desktop history limit. Downloaded comments are retained.".into());
                }
                let page_comments = data["comments"]
                    .as_array()
                    .ok_or("Incomplete comment response")?;
                comments.extend(page_comments.clone());
                let mut replacement = super::model::empty_workspace(workspace);
                replacement["comments"] = if finished {
                    json!(comments)
                } else {
                    json!(page_comments)
                };
                replacement["commentsCompleteFor"] = if finished { json!([id]) } else { json!([]) };
                replacement["syncedAt"] = snapshot["syncedAt"].clone();
                runtime.store.snapshot(account, workspace, &replacement)?;
                finished
            };
            if finished {
                break;
            }
            if page == 100 {
                return Err("The first 5,000 comments are downloaded. Open the web app for the full discussion.".into());
            }
            tokio::time::sleep(Duration::from_millis(25)).await;
        }
    }
    Ok(())
}

#[tauri::command]
pub async fn desktop_refresh_resource(
    window: tauri::WebviewWindow,
    workspace_id: String,
    resource_id: String,
    kind: String,
    expected_account_id: String,
) -> Result<Value, String> {
    let runtime = super::runtime(&window)?;
    let account = super::require_account(&runtime, &expected_account_id)?;
    uuid::Uuid::parse_str(&resource_id).map_err(|_| "Invalid linked item")?;
    let parameter = match kind.as_str() {
        "story" => "storyIds",
        "document" => "documentIds",
        _ => return Err("Unsupported resource".into()),
    };
    let snapshot = {
        let _guard = runtime.sync.lock().await;
        super::require_account(&runtime, &expected_account_id)?;
        ensure_current(&runtime, &account, &workspace_id)?;
        fetch(
            &runtime,
            &account,
            &workspace_id,
            &[("scope", "selected".into()), (parameter, resource_id)],
        )
        .await?
    };
    complete_threads(&runtime, &account, &workspace_id, &snapshot).await?;
    super::require_account(&runtime, &expected_account_id)?;
    let _ = window.emit("desktop-state-changed", ());
    runtime.store.workspace(&account, &workspace_id)
}

#[derive(Debug, PartialEq)]
pub(super) enum WarmOutcome {
    Complete,
    Yielded,
}

fn next_offset(page: &Value, offset: u64) -> Result<Option<u64>, String> {
    let next = page
        .get("nextStoryOffset")
        .ok_or("Incomplete download pagination")?;
    if next.is_null() {
        return Ok(None);
    }
    let next = next
        .as_u64()
        .filter(|next| *next > offset)
        .ok_or("The download did not advance; retry to continue")?;
    Ok(Some(next))
}

pub(super) async fn warm(
    app: &AppHandle,
    runtime: &Runtime,
    workspace: &str,
) -> Result<WarmOutcome, String> {
    let account = runtime.store.account()?;
    if !is_current(runtime, &account, workspace) || pending_changes(runtime, &account)? {
        return Ok(WarmOutcome::Yielded);
    }
    progress(app, runtime, &account, workspace, "settings", 0, None)?;
    settings::synchronize(runtime, workspace).await?;
    let mut downloaded = HashSet::new();
    let mut offset = 0u64;
    let mut recent = true;
    loop {
        if !is_current(runtime, &account, workspace) || pending_changes(runtime, &account)? {
            return Ok(WarmOutcome::Yielded);
        }
        progress(
            app,
            runtime,
            &account,
            workspace,
            if recent { "recent" } else { "my-work" },
            downloaded.len(),
            None,
        )?;
        let page = {
            let _guard = runtime.sync.lock().await;
            if !is_current(runtime, &account, workspace) {
                return Ok(WarmOutcome::Yielded);
            }
            let query = if recent {
                vec![("storyLimit", "25".into()), ("documentLimit", "5".into())]
            } else {
                vec![
                    ("myWork", "true".into()),
                    ("storyOffset", offset.to_string()),
                    ("storyLimit", "50".into()),
                    ("documentLimit", "1".into()),
                ]
            };
            fetch(runtime, &account, workspace, &query).await?
        };
        if !is_current(runtime, &account, workspace) {
            return Ok(WarmOutcome::Yielded);
        }
        for collection in ["stories", "documents"] {
            for id in page[collection]
                .as_array()
                .into_iter()
                .flatten()
                .filter_map(|item| item["id"].as_str())
            {
                downloaded.insert(format!("{collection}:{id}"));
            }
        }
        // Snapshot responses include a bounded first comment page. Full history
        // is fetched only when a story is opened, never for every startup item.
        let _ = app.emit("desktop-state-changed", ());
        if recent {
            recent = false;
        } else if let Some(next) = next_offset(&page, offset)? {
            offset = next;
        } else {
            break;
        }
        tokio::time::sleep(Duration::from_millis(50)).await;
    }
    progress(
        app,
        runtime,
        &account,
        workspace,
        "ready",
        downloaded.len(),
        None,
    )?;
    Ok(WarmOutcome::Complete)
}

#[cfg(test)]
mod tests;
