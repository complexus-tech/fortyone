#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod chrome;
#[cfg(target_os = "macos")]
mod macos_chrome;
mod navigation;
#[cfg(any(feature = "production", test))]
mod session;
mod window_controls;

use std::sync::Arc;

use navigation::{can_open_externally, NavigationPolicy};
use tauri::{
    menu::{Menu, MenuItem, MenuItemKind, PredefinedMenuItem, Submenu},
    webview::{NewWindowResponse, PageLoadEvent},
    AppHandle, Manager, Url, WebviewUrl, WebviewWindowBuilder,
};
use tauri_plugin_opener::OpenerExt;

const WINDOW_LABEL: &str = "main";

fn open_in_browser(app: &AppHandle, url: &Url) {
    if !can_open_externally(url) {
        eprintln!("Blocked an unsupported external URL scheme or embedded credentials");
        return;
    }

    if let Err(error) = app.opener().open_url(url.as_str(), None::<&str>) {
        eprintln!("Could not open the link in your browser: {error}");
    }
}

fn prepare_app_destination(
    app: &AppHandle,
    policy: &NavigationPolicy,
    url: &Url,
) -> Result<Option<Url>, String> {
    let Some(destination) = policy.app_destination(url) else {
        return Ok(None);
    };
    app.state::<chrome::ChromeCapabilities>()
        .ensure(app, policy, &destination)?;
    Ok(Some(destination))
}

fn navigate_main_window(app: &AppHandle, url: Url) {
    let app = app.clone();
    // Queue from a worker so the current native navigation decision finishes
    // before asking the same webview to load its replacement destination.
    tauri::async_runtime::spawn(async move {
        let Some(window) = app.get_webview_window(WINDOW_LABEL) else {
            eprintln!("Could not find the main desktop window for navigation");
            return;
        };
        if let Err(error) = window.navigate(url) {
            eprintln!("Could not open the app destination: {error}");
        }
    });
}

fn main() {
    tauri::Builder::default()
        .manage(chrome::ChromeController::default())
        .manage(chrome::ChromeCapabilities::default())
        .invoke_handler(tauri::generate_handler![
            chrome::sync_desktop_chrome,
            window_controls::desktop_navigation_state,
            window_controls::desktop_navigate,
            window_controls::desktop_window_action,
        ])
        .plugin(
            tauri_plugin_opener::Builder::new()
                // Link handling stays in Rust; hosted pages have no opener permission.
                .open_js_links_on_click(false)
                .build(),
        )
        .plugin(tauri_plugin_window_state::Builder::default().build())
        .setup(|app| {
            #[cfg(feature = "production")]
            let policy = Arc::new(NavigationPolicy::production());
            #[cfg(not(feature = "production"))]
            let policy = {
                let configured_url = std::env::var("FORTYONE_DESKTOP_URL")
                    .unwrap_or_else(|_| env!("FORTYONE_DESKTOP_DEFAULT_URL").to_owned());
                Arc::new(NavigationPolicy::new(Url::parse(&configured_url)?)?)
            };
            app.manage(Arc::clone(&policy));
            #[cfg(feature = "production")]
            let start_url = session::start_url(app.handle(), &policy);
            #[cfg(not(feature = "production"))]
            let start_url = policy.app_url().clone();
            app.state::<chrome::ChromeCapabilities>().ensure(
                app.handle(),
                &policy,
                policy.app_url(),
            )?;
            app.state::<chrome::ChromeCapabilities>()
                .ensure(app.handle(), &policy, &start_url)?;

            let navigation_policy = Arc::clone(&policy);
            let new_window_policy = Arc::clone(&policy);
            let navigation_app = app.handle().clone();
            let new_window_app = app.handle().clone();

            let builder =
                WebviewWindowBuilder::new(app, WINDOW_LABEL, WebviewUrl::External(start_url))
                    .title("FortyOne")
                    .inner_size(1280.0, 800.0)
                    .min_inner_size(960.0, 640.0)
                    .center()
                    // Keep the browser's file drag-and-drop behavior used by the web app.
                    .disable_drag_drop_handler()
                    .on_navigation(move |url| {
                        #[cfg(feature = "production")]
                        session::clear_after_logout(&navigation_app, &navigation_policy, url);
                        match prepare_app_destination(&navigation_app, &navigation_policy, url) {
                            Ok(Some(destination)) if destination == *url => return true,
                            Ok(Some(destination)) => {
                                navigate_main_window(&navigation_app, destination)
                            }
                            Ok(None) => open_in_browser(&navigation_app, url),
                            Err(error) => eprintln!("Blocked desktop navigation: {error}"),
                        }
                        false
                    })
                    .on_new_window(move |url, _features| {
                        match prepare_app_destination(&new_window_app, &new_window_policy, &url) {
                            Ok(Some(destination)) => {
                                navigate_main_window(&new_window_app, destination)
                            }
                            Ok(None) => open_in_browser(&new_window_app, &url),
                            Err(error) => eprintln!("Blocked desktop navigation: {error}"),
                        }
                        NewWindowResponse::Deny
                    })
                    .on_page_load(|window, payload| {
                        if payload.event() == PageLoadEvent::Started {
                            if let Err(error) = chrome::reset(&window) {
                                eprintln!(
                                    "Could not reset desktop actions while the page loads: {error}"
                                );
                            }
                        }
                    });

            #[cfg(target_os = "macos")]
            let builder = builder
                .title_bar_style(tauri::TitleBarStyle::Overlay)
                .hidden_title(true)
                // One native material spans the header and sidebar. The web
                // app keeps its content canvas opaque above this underlay.
                .transparent(true)
                .effects(
                    tauri::window::EffectsBuilder::new()
                        .effect(tauri::window::Effect::Sidebar)
                        .state(tauri::window::EffectState::FollowsWindowActiveState)
                        .build(),
                )
                .initialization_script(include_str!("bridge.js"));
            let window = builder.build()?;
            #[cfg(feature = "production")]
            {
                let session_window = window.clone();
                window.on_window_event(move |event| {
                    if matches!(event, tauri::WindowEvent::CloseRequested { .. }) {
                        session::save_window(&session_window);
                    }
                });
            }
            #[cfg(target_os = "macos")]
            {
                macos_chrome::install(&window)?;
                let chrome_window = window.clone();
                window.on_window_event(move |event| {
                    if matches!(
                        event,
                        tauri::WindowEvent::Resized(_)
                            | tauri::WindowEvent::ScaleFactorChanged { .. }
                            | tauri::WindowEvent::Focused(true)
                    ) {
                        let window = chrome_window.clone();
                        // Defer until AppKit finishes the native layout pass.
                        tauri::async_runtime::spawn(async move {
                            if let Err(error) = macos_chrome::refresh_layout(&window) {
                                eprintln!("Could not align macOS window controls: {error}");
                            }
                        });
                    }
                });
            }

            let menu = Menu::default(app.handle())?;
            #[cfg(target_os = "macos")]
            for item in menu.items()? {
                if let MenuItemKind::Submenu(submenu) = item {
                    match submenu.text()?.as_str() {
                        "File" => submenu.insert_items(
                            &[
                                &MenuItem::with_id(
                                    app,
                                    "create-story",
                                    "New Story…",
                                    false,
                                    Some("CmdOrCtrl+N"),
                                )?,
                                &PredefinedMenuItem::separator(app)?,
                            ],
                            0,
                        )?,
                        "View" => submenu.insert_items(
                            &[
                                &MenuItem::with_id(
                                    app,
                                    "search",
                                    "Search…",
                                    false,
                                    Some("CmdOrCtrl+K"),
                                )?,
                                &PredefinedMenuItem::separator(app)?,
                            ],
                            0,
                        )?,
                        _ => {}
                    }
                }
            }
            let navigate_menu = Submenu::with_items(
                app.handle(),
                "Navigate",
                true,
                &[
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
            menu.append(&navigate_menu)?;
            app.set_menu(menu)?;

            app.on_menu_event(move |app, event| {
                let Some(window) = app.get_webview_window(WINDOW_LABEL) else {
                    return;
                };
                let result = match event.id().as_ref() {
                    "create-story" | "search" => chrome::dispatch(&window, event.id().as_ref()),
                    "back" => window.eval("window.history.back()"),
                    "forward" => window.eval("window.history.forward()"),
                    "reload" => window.eval("window.location.reload()"),
                    "home" => window.navigate(policy.app_url().clone()),
                    "browser" => window.url().map(|url| open_in_browser(app, &url)),
                    "devtools" => {
                        #[cfg(debug_assertions)]
                        window.open_devtools();
                        Ok(())
                    }
                    _ => Ok(()),
                };
                if let Err(error) = result {
                    eprintln!("The desktop navigation action failed: {error}");
                }
            });
            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("Could not start FortyOne desktop")
        .run(|_app, _event| {
            #[cfg(feature = "production")]
            if matches!(_event, tauri::RunEvent::ExitRequested { .. }) {
                if let Some(window) = _app.get_webview_window(WINDOW_LABEL) {
                    session::save_window(&window);
                }
            }
        });
}
