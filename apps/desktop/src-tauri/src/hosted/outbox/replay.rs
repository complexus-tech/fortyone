use super::*;

const MAX_RESPONSE: usize = 8 * 1024 * 1024;

async fn response_body(mut response: reqwest::Response) -> Result<Value, String> {
    if response
        .content_length()
        .is_some_and(|size| size > MAX_RESPONSE as u64)
    {
        return Err("The saved item exceeds the desktop response limit".into());
    }
    let mut bytes = Vec::new();
    while let Some(chunk) = response
        .chunk()
        .await
        .map_err(|_| "The response was interrupted. Retry to reconcile the same operation.")?
    {
        if bytes.len().saturating_add(chunk.len()) > MAX_RESPONSE {
            return Err("The saved item exceeds the desktop response limit".into());
        }
        bytes.extend_from_slice(&chunk);
    }
    serde_json::from_slice(&bytes).map_err(|_| {
        "The server response was incomplete. Retry to reconcile the same operation.".into()
    })
}

async fn verify(state: &State, scope: &BoundScope) -> Result<(), VerifyError> {
    let user = get(state, "/auth/me", &scope.token).await?;
    if user["id"].as_str() != Some(&scope.account) {
        return Err(VerifyError::Denied);
    }
    let workspace = get(state, &format!("/workspaces/{}", scope.slug), &scope.token).await?;
    if workspace["id"].as_str() != Some(&scope.workspace) {
        return Err(VerifyError::Denied);
    }
    Ok(())
}

async fn current(window: &WebviewWindow, state: &State, scope: &BoundScope) -> Result<(), String> {
    if fingerprint(&browser_token(window).await?) != scope.cookie_hash
        || caller_origin(window, &scope.slug)? != scope.origin
    {
        return Err("Your account or workspace changed while synchronizing".into());
    }
    let db = state.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, scope)
}

fn mark(state: &State, scope: &BoundScope, operation: &Operation) -> Result<(), String> {
    let db = state.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, scope)?;
    store::save(&db, scope, operation)
}

async fn conflict(state: &State, scope: &BoundScope, operation: &mut Operation) {
    if operation.kind != "story.update" {
        return;
    }
    let path = format!(
        "/workspaces/{}/desktop/sync?scope=selected&storyIds={}",
        scope.slug, operation.resource_id
    );
    if let Ok(snapshot) = get(state, &path, &scope.token).await {
        if let Some(story) = snapshot["stories"].as_array().and_then(|items| {
            items
                .iter()
                .find(|item| item["id"].as_str() == Some(&operation.resource_id))
        }) {
            operation.server_version = story["updatedAt"].as_str().map(str::to_owned);
            operation.server_value = Some(story.clone());
        }
    }
}

pub(super) async fn synchronize(
    window: &WebviewWindow,
    state: &State,
    scope: &BoundScope,
) -> Result<(), String> {
    match verify(state, scope).await {
        Ok(()) => {}
        Err(VerifyError::Unavailable) => {
            return Err(
                "Could not verify access online. Your changes remain saved on this device.".into(),
            )
        }
        Err(VerifyError::Denied) => {
            let db = state.db.lock().map_err(|error| error.to_string())?;
            ensure_scope_current(&db, scope)?;
            forget_origin(&db, &scope.origin)?;
            return Err("Sign in and verify workspace access before synchronizing. Your changes are retained.".into());
        }
    }
    current(window, state, scope).await?;
    loop {
        let operation = {
            let db = state.db.lock().map_err(|error| error.to_string())?;
            ensure_scope_current(&db, scope)?;
            next_operation(store::read(&db, scope)?.operations)
        };
        let Some(mut operation) = operation else {
            return Ok(());
        };
        current(window, state, scope).await?;
        let payload = if operation.kind == "attachment.upload" {
            let db = state.db.lock().map_err(|error| error.to_string())?;
            ensure_scope_current(&db, scope)?;
            match super::super::attachments::wire_payload(&db, scope, &operation) {
                Ok(payload) => payload,
                Err(error) => {
                    operation.status = "blocked".into();
                    operation.error = Some(error);
                    store::save(&db, scope, &operation)?;
                    continue;
                }
            }
        } else {
            operation.payload.clone()
        };
        operation.attempted = true;
        mark(state, scope, &operation)?;
        let response=state.client.post(format!("{}/workspaces/{}/desktop/operations",api_origin(),scope.slug))
            .header(reqwest::header::COOKIE,format!("fortyone_session={}",scope.token))
            .header(reqwest::header::ORIGIN,&scope.origin)
            .header(reqwest::header::ACCEPT,"application/json")
            .json(&json!({"id":operation.id,"kind":operation.kind,"resourceId":operation.resource_id,"payload":payload,"baseVersion":operation.base_version}))
            .send().await;
        let response = match response {
            Ok(response) => response,
            Err(_) => {
                operation.error=Some("Connection interrupted. This change will be retried with the same operation ID.".into());
                mark(state, scope, &operation)?;
                return Err("You are offline. Changes are saved on this device.".into());
            }
        };
        let status = response.status().as_u16();
        let body = response_body(response).await;
        current(window, state, scope).await?;
        if (200..300).contains(&status) {
            let result = body?;
            let result = result.get("data").unwrap_or(&result);
            let mut db = state.db.lock().map_err(|error| error.to_string())?;
            let tx = db.transaction().map_err(|error| error.to_string())?;
            ensure_scope_current(&tx, scope)?;
            store::acknowledge(&tx, scope, &operation, result)?;
            tx.commit().map_err(|error| error.to_string())?;
            continue;
        }
        let message = body
            .ok()
            .and_then(|value| {
                value["error"]["message"]
                    .as_str()
                    .or_else(|| value["error"].as_str())
                    .or_else(|| value["message"].as_str())
                    .map(str::to_owned)
            })
            .unwrap_or_else(|| format!("The server could not save this change (HTTP {status})"));
        operation.status = match status {
            409 | 412 => "conflict",
            400 | 401 | 403 | 404 | 413 | 422 => "blocked",
            _ => "pending",
        }
        .into();
        operation.error = Some(message.clone());
        if operation.status == "conflict" {
            conflict(state, scope, &mut operation).await;
        }
        mark(state, scope, &operation)?;
        if status == 401 || status == 403 {
            let db = state.db.lock().map_err(|error| error.to_string())?;
            ensure_scope_current(&db, scope)?;
            forget_origin(&db, &scope.origin)?;
            return Err("Verify your account and workspace access before retrying. Local changes are retained.".into());
        }
        if operation.status == "pending" {
            return Err(message);
        }
        // Known rejection blocks only dependent work; independent changes can
        // continue through the same durable queue.
    }
}
