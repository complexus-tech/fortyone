use std::sync::Arc;

use serde::{Deserialize, Serialize};
use tauri::{menu::MenuItemKind, Manager, State, WebviewWindow};

use crate::{chrome::authorize_window, navigation::NavigationPolicy};

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct NavigationState {
    can_go_back: bool,
    can_go_forward: bool,
}

#[derive(Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum NavigationDirection {
    Back,
    Forward,
}

#[derive(Deserialize)]
#[serde(rename_all = "kebab-case")]
pub enum WindowAction {
    Drag,
    ToggleMaximize,
}

#[tauri::command]
pub async fn desktop_navigation_state(
    window: WebviewWindow,
    policy: State<'_, Arc<NavigationPolicy>>,
) -> Result<NavigationState, String> {
    authorize_window(&window, &policy)?;
    #[cfg(target_os = "macos")]
    let (can_go_back, can_go_forward) = crate::macos_chrome::navigation_state(&window).await?;
    #[cfg(not(target_os = "macos"))]
    let (can_go_back, can_go_forward) = (false, false);

    if let Some(menu) = window.app_handle().menu() {
        for entry in menu.items().map_err(|error| error.to_string())? {
            if let MenuItemKind::Submenu(submenu) = entry {
                for (id, enabled) in [("back", can_go_back), ("forward", can_go_forward)] {
                    if let Some(MenuItemKind::MenuItem(item)) = submenu.get(id) {
                        item.set_enabled(enabled)
                            .map_err(|error| error.to_string())?;
                    }
                }
            }
        }
    }
    Ok(NavigationState {
        can_go_back,
        can_go_forward,
    })
}

#[tauri::command]
pub async fn desktop_navigate(
    window: WebviewWindow,
    policy: State<'_, Arc<NavigationPolicy>>,
    direction: NavigationDirection,
) -> Result<(), String> {
    authorize_window(&window, &policy)?;
    navigate(&window, direction).await
}

pub async fn navigate(
    window: &WebviewWindow,
    direction: NavigationDirection,
) -> Result<(), String> {
    #[cfg(target_os = "macos")]
    return crate::macos_chrome::navigate(window, matches!(direction, NavigationDirection::Back))
        .await;
    #[cfg(not(target_os = "macos"))]
    window
        .eval(match direction {
            NavigationDirection::Back => "window.history.back()",
            NavigationDirection::Forward => "window.history.forward()",
        })
        .map_err(|error| error.to_string())
}

#[tauri::command]
pub fn desktop_window_action(
    window: WebviewWindow,
    policy: State<'_, Arc<NavigationPolicy>>,
    action: WindowAction,
) -> Result<(), String> {
    authorize_window(&window, &policy)?;
    match action {
        WindowAction::Drag => window.start_dragging(),
        WindowAction::ToggleMaximize => {
            if window.is_maximized().map_err(|error| error.to_string())? {
                window.unmaximize()
            } else {
                window.maximize()
            }
        }
    }
    .map_err(|error| error.to_string())
}
