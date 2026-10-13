use std::{collections::HashSet, sync::Mutex};

use serde::Deserialize;
use tauri::{
    ipc::CapabilityBuilder, menu::MenuItemKind, AppHandle, Manager, State, Url, WebviewWindow,
};

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
        .permission("allow-sync-desktop-chrome")
        .permission("allow-desktop-navigation-state")
        .permission("allow-desktop-navigate")
        .permission("allow-desktop-window-action")
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
    if let Some(menu) = window.app_handle().menu() {
        for entry in menu.items()? {
            if let MenuItemKind::Submenu(submenu) = entry {
                if let Some(MenuItemKind::MenuItem(item)) = submenu.get("create-story") {
                    item.set_text(format!("New {}…", state.story_label))?;
                    item.set_enabled(state.can_create_story && state.ready)?;
                }
                if let Some(MenuItemKind::MenuItem(item)) = submenu.get("search") {
                    item.set_enabled(state.ready)?;
                }
            }
        }
    }
    #[cfg(target_os = "macos")]
    crate::macos_chrome::update(window, state)?;
    Ok(())
}

pub fn reset(window: &WebviewWindow) -> tauri::Result<()> {
    let controller = window.state::<ChromeController>();
    let mut state = controller
        .0
        .lock()
        .unwrap_or_else(|error| error.into_inner());
    state.ready = false;
    state.can_create_story = false;
    apply(window, state.clone())
}

pub fn dispatch(window: &WebviewWindow, action: &str) -> tauri::Result<()> {
    let controller = window.state::<ChromeController>();
    let state = controller
        .0
        .lock()
        .unwrap_or_else(|error| error.into_inner());
    if !state.ready || (action == "create-story" && !state.can_create_story) {
        return Ok(());
    }
    // These fixed scripts only notify React; existing dialogs retain all mutation logic.
    let script = match action {
        "create-story" => "window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'create-story'}}))",
        "search" => "window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'search'}}))",
        _ => return Ok(()),
    };
    window.eval(script)
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
        assert_eq!(capability.permissions.len(), 4);
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
        assert_eq!(capability.permissions.len(), 4);
        assert_eq!(
            capability
                .permissions
                .iter()
                .map(|permission| permission.identifier().as_ref())
                .collect::<Vec<_>>(),
            vec![
                "allow-sync-desktop-chrome",
                "allow-desktop-navigation-state",
                "allow-desktop-navigate",
                "allow-desktop-window-action"
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
                    assert_eq!(capability.permissions.len(), 4);
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
