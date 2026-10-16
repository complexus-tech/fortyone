use std::sync::{Arc, Mutex};

use serde::Deserialize;
use tauri::{
    menu::{Menu, MenuItem, MenuItemKind, PredefinedMenuItem, Submenu},
    AppHandle, Manager, State, Url, WebviewWindow,
};

use crate::{
    chrome::{authorize_window, ChromeController, ChromeState},
    navigation::NavigationPolicy,
    window_controls::{navigate, NavigationDirection},
    WINDOW_LABEL,
};

const DOCUMENTATION_URL: &str = "https://docs.fortyone.app";
const SUPPORT_URL: &str = "mailto:info@fortyone.app";
const ACTIONS: [&str; 8] = [
    "settings",
    "logout",
    "create-story",
    "create-objective",
    "search",
    "toggle-sidebar",
    "keyboard-shortcuts",
    "product-tour",
];

/// Menu presentation only. The existing React actions retain all authorization
/// and mutation behavior; arbitrary scripts, destinations, and menu IDs are not accepted.
#[derive(Clone, Debug, Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct DesktopMenuState {
    pub ready: bool,
    pub can_create_story: bool,
    pub can_create_objective: bool,
    pub can_toggle_sidebar: bool,
    pub sidebar_collapsed: bool,
    pub story_label: String,
    pub objective_label: String,
}

impl Default for DesktopMenuState {
    fn default() -> Self {
        Self {
            ready: false,
            can_create_story: false,
            can_create_objective: false,
            can_toggle_sidebar: false,
            sidebar_collapsed: false,
            story_label: "Story".into(),
            objective_label: "Objective".into(),
        }
    }
}

impl DesktopMenuState {
    fn validate(&mut self) -> Result<(), String> {
        for label in [&mut self.story_label, &mut self.objective_label] {
            *label = label.trim().to_owned();
            if label.is_empty() || label.chars().count() > 40 || label.chars().any(char::is_control)
            {
                return Err("Desktop menu labels must contain 1–40 printable characters".into());
            }
        }
        self.can_create_story &= self.ready;
        self.can_create_objective &= self.ready;
        self.can_toggle_sidebar &= self.ready;
        Ok(())
    }

    fn allows(&self, action: &str) -> bool {
        self.ready
            && match action {
                "create-story" => self.can_create_story,
                "create-objective" => self.can_create_objective,
                "toggle-sidebar" => self.can_toggle_sidebar,
                "settings" | "logout" | "search" | "keyboard-shortcuts" | "product-tour" => true,
                _ => false,
            }
    }

    fn sidebar_label(&self) -> &'static str {
        if self.sidebar_collapsed {
            "Expand Sidebar"
        } else {
            "Collapse Sidebar"
        }
    }
}

/// None means the hosted frontend predates the optional menu bridge. Only its
/// established New Story and Search actions may fall back to chrome readiness.
#[derive(Default)]
pub struct MenuController(pub Mutex<Option<DesktopMenuState>>);

fn action_enabled(state: Option<&DesktopMenuState>, legacy: &ChromeState, action: &str) -> bool {
    match state {
        Some(state) => state.allows(action),
        None => {
            legacy.ready
                && match action {
                    "create-story" => legacy.can_create_story,
                    "search" => true,
                    _ => false,
                }
        }
    }
}

#[tauri::command]
pub fn sync_desktop_menu(
    window: WebviewWindow,
    policy: State<'_, Arc<NavigationPolicy>>,
    controller: State<'_, MenuController>,
    mut state: DesktopMenuState,
) -> Result<(), String> {
    authorize_window(&window, &policy)?;
    state.validate()?;
    *controller.0.lock().map_err(|error| error.to_string())? = Some(state);
    refresh(&window).map_err(|error| error.to_string())
}

pub fn reset(window: &WebviewWindow) {
    *window
        .state::<MenuController>()
        .0
        .lock()
        .unwrap_or_else(|error| error.into_inner()) = None;
}

pub fn refresh(window: &WebviewWindow) -> tauri::Result<()> {
    let menu_state = window
        .state::<MenuController>()
        .0
        .lock()
        .unwrap_or_else(|error| error.into_inner())
        .clone();
    let legacy = window
        .state::<ChromeController>()
        .0
        .lock()
        .unwrap_or_else(|error| error.into_inner())
        .clone();
    let Some(menu) = window.app_handle().menu() else {
        return Ok(());
    };
    for entry in menu.items()? {
        let MenuItemKind::Submenu(submenu) = entry else {
            continue;
        };
        for action in ACTIONS {
            let Some(MenuItemKind::MenuItem(item)) = submenu.get(action) else {
                continue;
            };
            item.set_enabled(action_enabled(menu_state.as_ref(), &legacy, action))?;
            match action {
                "create-story" => item.set_text(format!(
                    "New {}…",
                    menu_state
                        .as_ref()
                        .map(|state| &state.story_label)
                        .unwrap_or(&legacy.story_label)
                ))?,
                "create-objective" => item.set_text(format!(
                    "New {}…",
                    menu_state
                        .as_ref()
                        .map(|state| state.objective_label.as_str())
                        .unwrap_or("Objective")
                ))?,
                "toggle-sidebar" => item.set_text(
                    menu_state
                        .as_ref()
                        .map(DesktopMenuState::sidebar_label)
                        .unwrap_or("Collapse Sidebar"),
                )?,
                _ => {}
            }
        }
    }
    Ok(())
}

fn action_script(action: &str) -> Option<&'static str> {
    // Never interpolate a hosted label, arbitrary command, or URL into JavaScript.
    match action {
        "settings" => Some("window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'settings'}}))"),
        "logout" => Some("window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'logout'}}))"),
        "create-story" => Some("window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'create-story'}}))"),
        "create-objective" => Some("window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'create-objective'}}))"),
        "search" => Some("window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'search'}}))"),
        "toggle-sidebar" => Some("window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'toggle-sidebar'}}))"),
        "product-tour" => Some("window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'product-tour'}}))"),
        "keyboard-shortcuts" => Some("window.dispatchEvent(new CustomEvent('fortyone:desktop-action', {detail: {action: 'keyboard-shortcuts'}}))"),
        _ => None,
    }
}

pub fn dispatch(window: &WebviewWindow, action: &str) -> tauri::Result<()> {
    let allowed = {
        let controller = window.state::<MenuController>();
        let state = controller
            .0
            .lock()
            .unwrap_or_else(|error| error.into_inner());
        let legacy = window.state::<ChromeController>();
        let legacy = legacy.0.lock().unwrap_or_else(|error| error.into_inner());
        action_enabled(state.as_ref(), &legacy, action)
    };
    if !allowed {
        return Ok(());
    }
    let Some(script) = action_script(action) else {
        return Ok(());
    };
    window.show()?;
    window.unminimize()?;
    window.set_focus()?;
    window.eval(script)
}

fn submenu(menu: &Menu<tauri::Wry>, label: &str) -> tauri::Result<Submenu<tauri::Wry>> {
    for item in menu.items()? {
        if let MenuItemKind::Submenu(submenu) = item {
            if submenu.text()? == label {
                return Ok(submenu);
            }
        }
    }
    let submenu = Submenu::new(menu.app_handle(), label, true)?;
    // Platforms without a default File/View menu still receive the app actions.
    let position = if label == "File" {
        0
    } else {
        menu.items()?.len().saturating_sub(2)
    };
    menu.insert(&submenu, position)?;
    Ok(submenu)
}

pub fn install(app: &AppHandle, policy: Arc<NavigationPolicy>) -> tauri::Result<()> {
    let menu = Menu::default(app)?;
    #[cfg(target_os = "macos")]
    if let Some(MenuItemKind::Submenu(application)) = menu.items()?.first() {
        application.insert_items(
            &[
                &MenuItem::with_id(app, "settings", "Settings…", false, Some("CmdOrCtrl+,"))?,
                &MenuItem::with_id(app, "logout", "Log Out…", false, None::<&str>)?,
                &PredefinedMenuItem::separator(app)?,
            ],
            2,
        )?;
    }
    let file = submenu(&menu, "File")?;
    file.insert_items(
        &[
            &MenuItem::with_id(
                app,
                "create-story",
                "New Story…",
                false,
                Some("CmdOrCtrl+N"),
            )?,
            &MenuItem::with_id(
                app,
                "create-objective",
                "New Objective…",
                false,
                Some("CmdOrCtrl+Shift+N"),
            )?,
            &PredefinedMenuItem::separator(app)?,
        ],
        0,
    )?;
    #[cfg(not(target_os = "macos"))]
    file.append_items(&[
        &PredefinedMenuItem::separator(app)?,
        &MenuItem::with_id(app, "settings", "Settings…", false, Some("CmdOrCtrl+,"))?,
        &MenuItem::with_id(app, "logout", "Log Out…", false, None::<&str>)?,
    ])?;
    submenu(&menu, "View")?.insert_items(
        &[
            &MenuItem::with_id(
                app,
                "toggle-sidebar",
                "Collapse Sidebar",
                false,
                Some("CmdOrCtrl+Shift+L"),
            )?,
            &MenuItem::with_id(app, "search", "Search…", false, Some("CmdOrCtrl+K"))?,
            &PredefinedMenuItem::separator(app)?,
            &MenuItem::with_id(
                app,
                "back",
                "Back",
                !cfg!(target_os = "macos"),
                Some("CmdOrCtrl+["),
            )?,
            &MenuItem::with_id(
                app,
                "forward",
                "Forward",
                !cfg!(target_os = "macos"),
                Some("CmdOrCtrl+]"),
            )?,
            &MenuItem::with_id(app, "reload", "Reload", true, Some("CmdOrCtrl+R"))?,
            &PredefinedMenuItem::separator(app)?,
        ],
        0,
    )?;
    let navigate_menu = Submenu::with_items(
        app,
        "Navigate",
        true,
        &[
            &MenuItem::with_id(app, "home", "Open Start Page", true, None::<&str>)?,
            &MenuItem::with_id(app, "browser", "Open in Browser", true, None::<&str>)?,
        ],
    )?;
    #[cfg(debug_assertions)]
    navigate_menu.append(&MenuItem::with_id(
        app,
        "devtools",
        "Developer Tools",
        true,
        Some("CmdOrCtrl+Alt+I"),
    )?)?;
    menu.insert(&navigate_menu, menu.items()?.len().saturating_sub(2))?;
    submenu(&menu, "Help")?.append_items(&[
        &MenuItem::with_id(app, "documentation", "Documentation", true, None::<&str>)?,
        &MenuItem::with_id(
            app,
            "keyboard-shortcuts",
            "Keyboard Shortcuts",
            false,
            Some("CmdOrCtrl+/"),
        )?,
        &MenuItem::with_id(app, "product-tour", "Product Tour", false, None::<&str>)?,
        &PredefinedMenuItem::separator(app)?,
        &MenuItem::with_id(app, "support", "Contact Support…", true, None::<&str>)?,
    ])?;
    app.set_menu(menu)?;
    if let Some(window) = app.get_webview_window(WINDOW_LABEL) {
        refresh(&window)?;
    }
    app.on_menu_event(move |app, event| {
        let Some(window) = app.get_webview_window(WINDOW_LABEL) else {
            return;
        };
        let action = event.id().as_ref();
        let result = match action {
            action if ACTIONS.contains(&action) => dispatch(&window, action),
            "back" | "forward" => {
                let direction = if action == "back" {
                    NavigationDirection::Back
                } else {
                    NavigationDirection::Forward
                };
                tauri::async_runtime::spawn(async move {
                    if let Err(error) = navigate(&window, direction).await {
                        eprintln!("Could not navigate desktop history: {error}");
                    }
                });
                Ok(())
            }
            "reload" => window.eval("window.location.reload()"),
            "home" => window.navigate(policy.app_url().clone()),
            "browser" => window.url().map(|url| crate::open_in_browser(app, &url)),
            "documentation" | "support" => {
                let url = if action == "documentation" {
                    DOCUMENTATION_URL
                } else {
                    SUPPORT_URL
                };
                crate::open_in_browser(app, &Url::parse(url).expect("Fixed help URLs are valid"));
                Ok(())
            }
            "devtools" => {
                #[cfg(debug_assertions)]
                window.open_devtools();
                Ok(())
            }
            _ => Ok(()),
        };
        if let Err(error) = result {
            eprintln!("The desktop menu action failed: {error}");
        }
    });
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn ready() -> DesktopMenuState {
        DesktopMenuState {
            ready: true,
            can_create_story: true,
            can_create_objective: true,
            can_toggle_sidebar: true,
            ..Default::default()
        }
    }

    #[test]
    fn menu_readiness_is_independent_of_chrome_readiness_on_settings_pages() {
        let legacy = ChromeState::default();
        let state = ready();
        for action in ACTIONS {
            assert!(action_enabled(Some(&state), &legacy, action), "{action}");
        }
    }

    #[test]
    fn supplied_disabled_state_never_falls_back_to_legacy_chrome() {
        let legacy = ChromeState {
            ready: true,
            can_create_story: true,
            ..Default::default()
        };
        for action in ACTIONS {
            assert!(!action_enabled(
                Some(&DesktopMenuState::default()),
                &legacy,
                action
            ));
            assert_eq!(
                action_enabled(None, &legacy, action),
                matches!(action, "create-story" | "search")
            );
        }
        for action in ACTIONS {
            assert!(!action_enabled(None, &ChromeState::default(), action));
        }
    }

    #[test]
    fn creation_and_sidebar_permissions_are_enforced_for_every_dispatch() {
        let mut state = ready();
        state.can_create_story = false;
        state.can_create_objective = false;
        state.can_toggle_sidebar = false;
        for action in [
            "create-story",
            "create-objective",
            "toggle-sidebar",
            "unknown",
        ] {
            assert!(!state.allows(action));
        }
        for action in ["settings", "logout", "search", "keyboard-shortcuts"] {
            assert!(state.allows(action));
        }
        state.ready = false;
        state.can_create_story = true;
        state.can_create_objective = true;
        state.can_toggle_sidebar = true;
        state.validate().unwrap();
        assert!(
            !state.can_create_story && !state.can_create_objective && !state.can_toggle_sidebar
        );
        for action in ACTIONS {
            assert!(!state.allows(action));
        }
    }

    #[test]
    fn labels_are_trimmed_bounded_and_cannot_include_control_characters() {
        let mut state = ready();
        state.story_label = "  Ticket  ".into();
        state.objective_label = "  Goal  ".into();
        state.validate().unwrap();
        assert_eq!(state.story_label, "Ticket");
        assert_eq!(state.objective_label, "Goal");
        assert_eq!(state.sidebar_label(), "Collapse Sidebar");
        state.sidebar_collapsed = true;
        assert_eq!(state.sidebar_label(), "Expand Sidebar");
        for label in [
            "".to_owned(),
            "a".repeat(41),
            "Goal\nInjected".to_owned(),
            "Goal\u{7f}".to_owned(),
        ] {
            for story in [false, true] {
                let mut state = ready();
                if story {
                    state.story_label = label.clone();
                } else {
                    state.objective_label = label.clone();
                }
                assert!(state.validate().is_err());
            }
        }
    }

    #[test]
    fn bridge_rejects_unknown_fields_and_scripts_have_a_fixed_action_allowlist() {
        let mut payload = serde_json::json!({ "ready": true, "canCreateStory": true, "canCreateObjective": true, "canToggleSidebar": true, "sidebarCollapsed": false, "storyLabel": "Ticket", "objectiveLabel": "Goal" });
        assert!(serde_json::from_value::<DesktopMenuState>(payload.clone()).is_ok());
        payload["script"] = "malicious()".into();
        assert!(serde_json::from_value::<DesktopMenuState>(payload).is_err());
        for action in ACTIONS {
            assert!(action_script(action)
                .unwrap()
                .contains(&format!("action: '{action}'")));
        }
        for action in ["", "documentation", "settings'}));alert(1)//"] {
            assert!(action_script(action).is_none());
        }
        for url in [DOCUMENTATION_URL, SUPPORT_URL] {
            assert!(crate::navigation::can_open_externally(
                &Url::parse(url).unwrap()
            ));
        }
    }
}
