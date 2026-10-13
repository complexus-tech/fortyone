#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod chrome;
#[cfg(target_os = "macos")]
mod macos_chrome;
mod navigation;
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

fn main() {
    tauri::Builder::default()
        .manage(chrome::ChromeController::default())
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
            let configured_url = std::env::var("FORTYONE_DESKTOP_URL")
                .unwrap_or_else(|_| env!("FORTYONE_DESKTOP_DEFAULT_URL").to_owned());
            let policy = Arc::new(NavigationPolicy::new(Url::parse(&configured_url)?)?);
            app.manage(Arc::clone(&policy));
            app.add_capability(chrome::capability(&policy))?;

            let navigation_policy = Arc::clone(&policy);
            let navigation_app = app.handle().clone();
            let new_window_app = app.handle().clone();

            let builder = WebviewWindowBuilder::new(
                app,
                WINDOW_LABEL,
                WebviewUrl::External(policy.app_url().clone()),
            )
            .title("FortyOne")
            .inner_size(1280.0, 800.0)
            .min_inner_size(960.0, 640.0)
            .center()
            // Keep the browser's file drag-and-drop behavior used by the web app.
            .disable_drag_drop_handler()
            .on_navigation(move |url| {
                if navigation_policy.allows_navigation(url) {
                    return true;
                }
                open_in_browser(&navigation_app, url);
                false
            })
            .on_new_window(move |url, _features| {
                open_in_browser(&new_window_app, &url);
                NewWindowResponse::Deny
            })
            .on_page_load(|window, payload| {
                if payload.event() == PageLoadEvent::Started {
                    if let Err(error) = chrome::reset(&window) {
                        eprintln!("Could not reset desktop actions while the page loads: {error}");
                    }
                }
            });

            #[cfg(target_os = "macos")]
            let builder = builder
                .title_bar_style(tauri::TitleBarStyle::Overlay)
                .hidden_title(true)
                .initialization_script(include_str!("bridge.js"));
            let window = builder.build()?;
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
            menu.append(&Submenu::with_items(
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
                    &MenuItem::with_id(
                        app,
                        "devtools",
                        "Developer Tools",
                        true,
                        Some("CmdOrCtrl+Alt+I"),
                    )?,
                ],
            )?)?;
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
        .run(tauri::generate_context!())
        .expect("Could not start FortyOne desktop");
}
