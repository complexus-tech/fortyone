use std::{
    sync::{Arc, Mutex},
    time::{Duration, Instant, SystemTime, UNIX_EPOCH},
};

use serde::{Deserialize, Serialize};
use tauri::{State, Url, WebviewWindow};

use crate::{chrome::authorize_window, navigation::NavigationPolicy};

const TRANSITION_TIMEOUT: Duration = Duration::from_secs(20);

#[derive(Clone, Deserialize, Serialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct WorkspaceDestination {
    id: String,
    name: String,
    url: String,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub struct WorkspaceReady {
    id: String,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct WorkspaceSwitchState {
    #[serde(flatten)]
    destination: WorkspaceDestination,
    started_at: u64,
    failed: bool,
}

struct WorkspaceTransition {
    destination: WorkspaceDestination,
    target_url: Url,
    // Recovery stays native and never exposes a previous page's private URL to JS.
    source_url: Url,
    started: Instant,
    started_at: u64,
}

impl WorkspaceTransition {
    fn snapshot(&self) -> WorkspaceSwitchState {
        WorkspaceSwitchState {
            destination: self.destination.clone(),
            started_at: self.started_at,
            failed: self.started.elapsed() >= TRANSITION_TIMEOUT,
        }
    }

    fn restart(&mut self) {
        self.started = Instant::now();
        self.started_at = epoch_milliseconds();
    }

    fn is_ready(&self, workspace: &WorkspaceReady, current_url: &Url) -> bool {
        self.destination.id == workspace.id
            && self.target_url.origin() == current_url.origin()
            && self.target_url.path() == current_url.path()
    }
}

#[derive(Default)]
pub struct WorkspaceTransitions(Mutex<Option<WorkspaceTransition>>);

fn epoch_milliseconds() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_millis()
        .min(u64::MAX as u128) as u64
}

fn validate_destination(
    policy: &NavigationPolicy,
    source_url: &Url,
    mut destination: WorkspaceDestination,
) -> Result<(WorkspaceDestination, Url), String> {
    if destination.id.is_empty()
        || destination.id.len() > 128
        || !destination
            .id
            .bytes()
            .all(|character| character.is_ascii_alphanumeric() || matches!(character, b'-' | b'_'))
    {
        return Err("The workspace identity is invalid".into());
    }
    destination.name = destination.name.trim().to_owned();
    if destination.name.is_empty()
        || destination.name.chars().count() > 160
        || destination.name.chars().any(char::is_control)
        || destination.url.len() > 2_048
    {
        return Err("The workspace destination is invalid".into());
    }
    let target = source_url
        .join(&destination.url)
        .map_err(|_| "The workspace URL is invalid")?;
    if !policy.allows_navigation(&target) || target.query().is_some() || target.fragment().is_some()
    {
        return Err("This workspace cannot be opened in the desktop app".into());
    }
    destination.url = target.to_string();
    Ok((destination, target))
}

#[tauri::command]
pub fn begin_workspace_switch(
    window: WebviewWindow,
    policy: State<'_, Arc<NavigationPolicy>>,
    transitions: State<'_, WorkspaceTransitions>,
    destination: WorkspaceDestination,
) -> Result<(), String> {
    authorize_window(&window, &policy)?;
    let mut source_url = window.url().map_err(|error| error.to_string())?;
    let (destination, target_url) = validate_destination(&policy, &source_url, destination)?;
    source_url.set_query(None);
    source_url.set_fragment(None);
    let mut state = transitions.0.lock().map_err(|error| error.to_string())?;
    if let Some(active) = state.as_mut() {
        if active.destination.id == destination.id && active.target_url == target_url {
            if active.started.elapsed() >= TRANSITION_TIMEOUT {
                active.restart();
            }
            return Ok(());
        }
        return Err("Finish the current workspace switch before opening another workspace".into());
    }
    *state = Some(WorkspaceTransition {
        destination,
        target_url,
        source_url,
        started: Instant::now(),
        started_at: epoch_milliseconds(),
    });
    Ok(())
}

#[tauri::command]
pub fn get_workspace_switch(
    window: WebviewWindow,
    policy: State<'_, Arc<NavigationPolicy>>,
    transitions: State<'_, WorkspaceTransitions>,
) -> Result<Option<WorkspaceSwitchState>, String> {
    authorize_window(&window, &policy)?;
    let state = transitions.0.lock().map_err(|error| error.to_string())?;
    Ok(state.as_ref().map(WorkspaceTransition::snapshot))
}

#[tauri::command]
pub fn workspace_ready(
    window: WebviewWindow,
    policy: State<'_, Arc<NavigationPolicy>>,
    transitions: State<'_, WorkspaceTransitions>,
    workspace: WorkspaceReady,
) -> Result<(), String> {
    authorize_window(&window, &policy)?;
    let current_url = window.url().map_err(|error| error.to_string())?;
    let mut state = transitions.0.lock().map_err(|error| error.to_string())?;
    if state
        .as_ref()
        .is_some_and(|active| active.is_ready(&workspace, &current_url))
    {
        *state = None;
    }
    Ok(())
}

#[tauri::command]
pub fn retry_workspace_switch(
    window: WebviewWindow,
    policy: State<'_, Arc<NavigationPolicy>>,
    transitions: State<'_, WorkspaceTransitions>,
) -> Result<(), String> {
    authorize_window(&window, &policy)?;
    let (target, started) = {
        let state = transitions.0.lock().map_err(|error| error.to_string())?;
        let active = state
            .as_ref()
            .ok_or("There is no workspace switch to retry")?;
        (active.target_url.clone(), active.started)
    };
    window
        .navigate(target.clone())
        .map_err(|error| error.to_string())?;
    let mut state = transitions.0.lock().map_err(|error| error.to_string())?;
    if let Some(active) = state.as_mut() {
        if active.target_url == target && active.started == started {
            active.restart();
        }
    }
    Ok(())
}

#[tauri::command]
pub fn cancel_workspace_switch(
    window: WebviewWindow,
    policy: State<'_, Arc<NavigationPolicy>>,
    transitions: State<'_, WorkspaceTransitions>,
) -> Result<(), String> {
    authorize_window(&window, &policy)?;
    let (source, started) = {
        let state = transitions.0.lock().map_err(|error| error.to_string())?;
        let Some(active) = state.as_ref() else {
            return Ok(());
        };
        (active.source_url.clone(), active.started)
    };
    window
        .eval("window.stop()")
        .map_err(|error| error.to_string())?;
    if window.url().map_err(|error| error.to_string())? != source {
        window
            .navigate(source.clone())
            .map_err(|error| error.to_string())?;
    }
    // Keep recovery available if stopping or returning failed, and never clear
    // a newer transition that was started while the native navigation ran.
    let mut state = transitions.0.lock().map_err(|error| error.to_string())?;
    if state
        .as_ref()
        .is_some_and(|active| active.source_url == source && active.started == started)
    {
        *state = None;
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn destination(url: &str) -> WorkspaceDestination {
        WorkspaceDestination {
            id: "workspace-two".into(),
            name: "Product".into(),
            url: url.into(),
        }
    }

    #[test]
    fn rejects_external_services_credentials_and_transient_target_state() {
        let policy = NavigationPolicy::production();
        let source = Url::parse("https://acme.fortyone.app/my-work").unwrap();
        for target in [
            "https://attacker.example/my-work",
            "https://api.fortyone.app/my-work",
            "http://product.fortyone.app/my-work",
            "https://user@product.fortyone.app/my-work",
            "https://product.fortyone.app/my-work?token=secret",
            "https://product.fortyone.app/my-work#fragment",
        ] {
            assert!(
                validate_destination(&policy, &source, destination(target)).is_err(),
                "{target}"
            );
        }
    }

    #[test]
    fn normalizes_local_targets_and_validates_display_values() {
        let source = Url::parse("http://localhost:3000/acme/my-work").unwrap();
        let policy = NavigationPolicy::new(source.clone()).unwrap();
        let (valid, url) =
            validate_destination(&policy, &source, destination("/product/my-work")).unwrap();
        assert_eq!(valid.url, "http://localhost:3000/product/my-work");
        assert_eq!(url.as_str(), valid.url);
        let mut invalid = destination("/product/my-work");
        invalid.name = "Product\nNotification".into();
        assert!(validate_destination(&policy, &source, invalid).is_err());
    }

    #[test]
    fn readiness_requires_target_identity_origin_and_route() {
        let transition = WorkspaceTransition {
            destination: destination("https://product.fortyone.app/my-work"),
            target_url: Url::parse("https://product.fortyone.app/my-work").unwrap(),
            source_url: Url::parse("https://acme.fortyone.app/my-work").unwrap(),
            started: Instant::now(),
            started_at: 123,
        };
        let ready = WorkspaceReady {
            id: "workspace-two".into(),
        };
        assert!(transition.is_ready(&ready, &transition.target_url));
        assert!(!transition.is_ready(&ready, &transition.source_url));
        assert!(!transition.is_ready(
            &ready,
            &Url::parse("https://product.fortyone.app/auth-callback").unwrap()
        ));
        assert!(!transition.is_ready(
            &WorkspaceReady {
                id: "workspace-one".into()
            },
            &transition.target_url
        ));
    }

    #[test]
    fn timeout_survives_documents_and_snapshot_never_exposes_source() {
        let mut transition = WorkspaceTransition {
            destination: destination("https://product.fortyone.app/my-work"),
            target_url: Url::parse("https://product.fortyone.app/my-work").unwrap(),
            source_url: Url::parse("https://private.fortyone.app/docs").unwrap(),
            started: Instant::now() - TRANSITION_TIMEOUT,
            started_at: 123,
        };
        assert!(transition.snapshot().failed);
        transition.restart();
        assert!(!transition.snapshot().failed);
        assert_eq!(
            transition.snapshot().destination.url,
            transition.target_url.as_str()
        );
        assert_eq!(
            transition.source_url.host_str(),
            Some("private.fortyone.app")
        );
    }
}
