use std::{collections::HashSet, sync::Mutex};

use serde::Deserialize;
use tauri::{ipc::CapabilityBuilder, AppHandle, Manager, State, Url, WebviewWindow};

use crate::{navigation::NavigationPolicy, WINDOW_LABEL};

/// Presentation state only: this bridge cannot read files, open URLs, or write app data.
#[derive(Clone, Debug, Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ChromeState {
    pub background: [u8; 3],
    pub dark: bool,
    pub system_theme: bool,
    pub ready: bool,
    pub can_create_story: bool,
    pub story_label: String,
}

impl Default for ChromeState {
    fn default() -> Self {
        Self {
            background: [255, 255, 255],
            dark: false,
            system_theme: true,
            ready: false,
            can_create_story: false,
            story_label: "story".into(),
        }
    }
}

impl ChromeState {
    fn validate(&mut self) -> Result<(), String> {
        self.story_label = self.story_label.trim().to_owned();
        if self.story_label.is_empty()
            || self.story_label.chars().count() > 40
            || self.story_label.chars().any(char::is_control)
        {
            return Err("The desktop story label must contain 1–40 printable characters".into());
        }
        self.can_create_story &= self.ready;
        Ok(())
    }
}

#[derive(Default)]
pub struct ChromeController(pub Mutex<ChromeState>);

const MAX_CAPABILITY_ORIGINS: usize = 256;

#[derive(Default)]
pub struct ChromeCapabilities(Mutex<HashSet<String>>);

impl ChromeCapabilities {
    pub fn ensure(
        &self,
        app: &AppHandle,
        policy: &NavigationPolicy,
        url: &Url,
    ) -> Result<(), String> {
        self.register(policy, url, |capability| {
            app.add_capability(capability)
                .map_err(|error| error.to_string())
        })
    }

    fn register(
        &self,
        policy: &NavigationPolicy,
        url: &Url,
        add: impl FnOnce(CapabilityBuilder) -> Result<(), String>,
    ) -> Result<(), String> {
        if !policy.allows_navigation(url) {
            return Err("Desktop controls cannot be granted to this origin".into());
        }
        let origin = url.origin().ascii_serialization();
        let mut registered = self.0.lock().map_err(|error| error.to_string())?;
        if registered.contains(&origin) {
            return Ok(());
        }
        // Tauri appends dynamic ACL entries and has no public removal API.
        // Register once per origin, and bound growth for this application session.
        if registered.len() >= MAX_CAPABILITY_ORIGINS {
            return Err("Restart FortyOne before opening additional workspace origins".into());
        }
        let identifier = format!("desktop-chrome-{}", registered.len());
        add(capability(policy, url, &identifier))?;
        registered.insert(origin);
        Ok(())
    }
}

fn capability(policy: &NavigationPolicy, url: &Url, identifier: &str) -> CapabilityBuilder {
    CapabilityBuilder::new(identifier)
        // Tauri classifies the configured devUrl/frontendDist as local content.
        .local(url.origin() == policy.app_url().origin())
        .webview(WINDOW_LABEL)
        .remote(format!("{}/*", url.origin().ascii_serialization()))
        // The notification plugin's injected bootstrap reads permission on load.
        // Sending and requesting permission remain behind our checked commands.
        .permission("notification:allow-is-permission-granted")
        .permission("allow-sync-desktop-chrome")
        .permission("allow-sync-desktop-menu")
        .permission("allow-desktop-navigation-state")
        .permission("allow-desktop-navigate")
        .permission("allow-desktop-window-action")
        .permission("allow-begin-workspace-switch")
        .permission("allow-get-workspace-switch")
        .permission("allow-workspace-ready")
        .permission("allow-retry-workspace-switch")
        .permission("allow-cancel-workspace-switch")
        .permission("allow-hosted-bind-session")
        .permission("allow-hosted-restore-session")
        .permission("allow-hosted-cache-read")
        .permission("allow-hosted-cache-write")
        .permission("allow-hosted-cache-clear")
        .permission("allow-hosted-draft-read")
        .permission("allow-hosted-draft-write")
        .permission("allow-hosted-draft-delete")
        .permission("allow-hosted-get-preferences")
        .permission("allow-hosted-update-preferences")
        .permission("allow-hosted-request-notification-permission")
        .permission("allow-hosted-deliver-notification")
        .permission("allow-hosted-set-notification-badge")
        .permission("allow-hosted-page-read")
        .permission("allow-hosted-page-write")
        .permission("allow-hosted-enqueue-operation")
        .permission("allow-hosted-read-outbox")
        .permission("allow-hosted-sync-outbox")
        .permission("allow-hosted-resolve-operation")
        .permission("allow-hosted-confirm-close")
        .permission("allow-hosted-stage-attachment")
        .permission("allow-hosted-read-attachment")
        .permission("allow-hosted-export-attachment")
        .permission("allow-hosted-get-update-status")
        .permission("allow-hosted-check-update")
        .permission("allow-hosted-install-update")
        .permission("allow-hosted-export-recovery")
}

pub fn authorize_window(window: &WebviewWindow, policy: &NavigationPolicy) -> Result<(), String> {
    if window.label() != WINDOW_LABEL
        || !policy.allows_navigation(&window.url().map_err(|error| error.to_string())?)
    {
        return Err("Desktop controls are only available to the configured app window".into());
    }
    Ok(())
}

#[tauri::command]
pub fn sync_desktop_chrome(
    window: WebviewWindow,
    policy: State<'_, std::sync::Arc<NavigationPolicy>>,
    controller: State<'_, ChromeController>,
    mut state: ChromeState,
) -> Result<(), String> {
    authorize_window(&window, &policy)?;
    state.validate()?;
    *controller.0.lock().map_err(|error| error.to_string())? = state.clone();
    apply(&window, state).map_err(|error| error.to_string())
}

pub fn apply(window: &WebviewWindow, state: ChromeState) -> tauri::Result<()> {
    crate::menus::refresh(window)?;
    #[cfg(target_os = "macos")]
    crate::macos_chrome::update(window, state)?;
    Ok(())
}

pub fn reset(window: &WebviewWindow) -> tauri::Result<()> {
    let state = {
        let controller = window.state::<ChromeController>();
        let mut state = controller
            .0
            .lock()
            .unwrap_or_else(|error| error.into_inner());
        state.ready = false;
        state.can_create_story = false;
        state.clone()
    };
    crate::menus::reset(window);
    apply(window, state)
}

pub fn dispatch(window: &WebviewWindow, action: &str) -> tauri::Result<()> {
    crate::menus::dispatch(window, action)
}

#[cfg(test)]
mod tests {
    use super::*;
    use tauri::{ipc::RuntimeCapability, utils::acl::capability::CapabilityFile, Url};

    #[test]
    fn production_grants_desktop_controls_only_to_the_production_webview() {
        let policy = NavigationPolicy::production();
        let CapabilityFile::Capability(capability) =
            capability(&policy, policy.app_url(), "desktop-chrome").build()
        else {
            panic!("Expected a single desktop capability");
        };
        assert_eq!(capability.webviews, vec![WINDOW_LABEL]);
        assert!(capability.windows.is_empty());
        assert_eq!(
            capability.remote.unwrap().urls,
            vec!["https://cloud.fortyone.app/*"]
        );
        assert_eq!(capability.permissions.len(), 38);
    }

    #[test]
    fn grants_only_desktop_controls_to_the_configured_app_origin() {
        let policy =
            NavigationPolicy::new(Url::parse("http://localhost:3456/acme/maya?view=all").unwrap())
                .unwrap();
        let CapabilityFile::Capability(capability) =
            capability(&policy, policy.app_url(), "desktop-chrome").build()
        else {
            panic!("Expected a single desktop capability");
        };
        assert!(capability.local);
        assert!(capability.windows.is_empty());
        assert_eq!(capability.webviews, vec![WINDOW_LABEL]);
        assert_eq!(
            capability.remote.unwrap().urls,
            vec!["http://localhost:3456/*"]
        );
        assert_eq!(capability.permissions.len(), 38);
        assert_eq!(
            capability
                .permissions
                .iter()
                .map(|permission| permission.identifier().as_ref())
                .collect::<Vec<_>>(),
            vec![
                "notification:allow-is-permission-granted",
                "allow-sync-desktop-chrome",
                "allow-sync-desktop-menu",
                "allow-desktop-navigation-state",
                "allow-desktop-navigate",
                "allow-desktop-window-action",
                "allow-begin-workspace-switch",
                "allow-get-workspace-switch",
                "allow-workspace-ready",
                "allow-retry-workspace-switch",
                "allow-cancel-workspace-switch",
                "allow-hosted-bind-session",
                "allow-hosted-restore-session",
                "allow-hosted-cache-read",
                "allow-hosted-cache-write",
                "allow-hosted-cache-clear",
                "allow-hosted-draft-read",
                "allow-hosted-draft-write",
                "allow-hosted-draft-delete",
                "allow-hosted-get-preferences",
                "allow-hosted-update-preferences",
                "allow-hosted-request-notification-permission",
                "allow-hosted-deliver-notification",
                "allow-hosted-set-notification-badge",
                "allow-hosted-page-read",
                "allow-hosted-page-write",
                "allow-hosted-enqueue-operation",
                "allow-hosted-read-outbox",
                "allow-hosted-sync-outbox",
                "allow-hosted-resolve-operation",
                "allow-hosted-confirm-close",
                "allow-hosted-stage-attachment",
                "allow-hosted-read-attachment",
                "allow-hosted-export-attachment",
                "allow-hosted-get-update-status",
                "allow-hosted-check-update",
                "allow-hosted-install-update",
                "allow-hosted-export-recovery",
            ]
        );
    }

    #[test]
    fn registers_exact_workspace_origins_once_across_navigation_and_new_windows() {
        let registry = ChromeCapabilities::default();
        let policy = NavigationPolicy::production();
        let mut grants = Vec::new();
        for target in [
            "https://cloud.fortyone.app/",
            "https://acme.fortyone.app/my-work",
            "https://other-team.fortyone.app/maya",
            "https://acme.fortyone.app/settings#profile",
            "https://cloud.fortyone.app:443/auth-callback",
        ] {
            registry
                .register(&policy, &Url::parse(target).unwrap(), |builder| {
                    let CapabilityFile::Capability(capability) = builder.build() else {
                        panic!("Expected one capability");
                    };
                    assert_eq!(capability.webviews, vec![WINDOW_LABEL]);
                    assert_eq!(capability.permissions.len(), 38);
                    assert_eq!(
                        capability
                            .permissions
                            .iter()
                            .map(|permission| permission.identifier().as_ref())
                            .filter(|permission| permission.starts_with("notification:"))
                            .collect::<Vec<_>>(),
                        vec!["notification:allow-is-permission-granted"]
                    );
                    grants.push((
                        capability.identifier,
                        capability.local,
                        capability.remote.unwrap().urls,
                    ));
                    Ok(())
                })
                .unwrap();
        }
        assert_eq!(
            grants,
            vec![
                (
                    "desktop-chrome-0".into(),
                    true,
                    vec!["https://cloud.fortyone.app/*".into()]
                ),
                (
                    "desktop-chrome-1".into(),
                    false,
                    vec!["https://acme.fortyone.app/*".into()]
                ),
                (
                    "desktop-chrome-2".into(),
                    false,
                    vec!["https://other-team.fortyone.app/*".into()]
                ),
            ]
        );
    }

    #[test]
    fn refuses_untrusted_capabilities_and_retries_failed_registration() {
        let registry = ChromeCapabilities::default();
        let policy = NavigationPolicy::production();
        for target in [
            "https://api.fortyone.app/",
            "https://one.two.fortyone.app/",
            "http://acme.fortyone.app/",
        ] {
            assert!(registry
                .register(&policy, &Url::parse(target).unwrap(), |_| {
                    panic!("Invalid origins must never reach capability registration");
                })
                .is_err());
        }
        let target = Url::parse("https://acme.fortyone.app/").unwrap();
        assert!(registry
            .register(&policy, &target, |_| Err("registration failed".into()))
            .is_err());
        assert!(registry.0.lock().unwrap().is_empty());
        registry.register(&policy, &target, |_| Ok(())).unwrap();
        assert_eq!(registry.0.lock().unwrap().len(), 1);
    }

    #[test]
    fn bounds_dynamic_capability_growth_without_blocking_registered_origins() {
        let registry = ChromeCapabilities::default();
        let policy = NavigationPolicy::production();
        for index in 0..MAX_CAPABILITY_ORIGINS {
            let url = Url::parse(&format!("https://workspace-{index}.fortyone.app/")).unwrap();
            registry.register(&policy, &url, |_| Ok(())).unwrap();
        }
        let new_origin = Url::parse("https://additional.fortyone.app/").unwrap();
        assert!(registry.register(&policy, &new_origin, |_| Ok(())).is_err());
        let registered = Url::parse("https://workspace-0.fortyone.app/settings").unwrap();
        registry
            .register(&policy, &registered, |_| {
                panic!("Must reuse the existing capability")
            })
            .unwrap();
    }

    #[test]
    fn loading_and_signed_out_pages_cannot_enable_creation() {
        let mut state = ChromeState {
            can_create_story: true,
            ..ChromeState::default()
        };
        state.validate().unwrap();
        assert!(!state.can_create_story);
    }

    #[test]
    fn rejects_invalid_native_labels() {
        for label in ["".to_owned(), "a".repeat(41), "task\nmenu".to_owned()] {
            let mut state = ChromeState {
                story_label: label,
                ..ChromeState::default()
            };
            assert!(state.validate().is_err());
        }
    }
}
