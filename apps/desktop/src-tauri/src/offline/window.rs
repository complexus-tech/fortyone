use tauri::{
    menu::{Menu, MenuItem},
    tray::TrayIconBuilder,
    AppHandle, Emitter, Manager, WebviewWindow,
};

pub fn show(app: &AppHandle) {
    if let Some(window) = app.get_webview_window("main") {
        let _ = window.show();
        let _ = window.unminimize();
        let _ = window.set_focus();
    }
}

pub fn install(app: &AppHandle, window: &WebviewWindow) -> tauri::Result<()> {
    let clone = window.clone();
    window.on_window_event(move |event| match event {
        tauri::WindowEvent::CloseRequested { api, .. } => {
            api.prevent_close();
            let _ = clone.hide();
        }
        #[cfg(target_os = "macos")]
        tauri::WindowEvent::Resized(_)
        | tauri::WindowEvent::ScaleFactorChanged { .. }
        | tauri::WindowEvent::Focused(true) => {
            let window = clone.clone();
            tauri::async_runtime::spawn(async move {
                if let Err(error) = crate::macos_chrome::refresh_layout(&window) {
                    eprintln!("Could not align window controls: {error}");
                }
            });
        }
        _ => {}
    });
    let menu = Menu::with_items(
        app,
        &[
            &MenuItem::with_id(app, "tray-open", "Open FortyOne", true, None::<&str>)?,
            &MenuItem::with_id(app, "tray-new", "New Story…", true, None::<&str>)?,
            &MenuItem::with_id(app, "tray-sync", "Sync Now", true, None::<&str>)?,
            &MenuItem::with_id(app, "tray-quit", "Quit FortyOne", true, None::<&str>)?,
        ],
    )?;
    let tray = TrayIconBuilder::with_id("fortyone")
        .tooltip("FortyOne")
        .menu(&menu)
        .on_menu_event(|app, event| match event.id().as_ref() {
            "tray-open" => show(app),
            "tray-new" => {
                show(app);
                let _ = app.emit(
                    "desktop-action",
                    serde_json::json!({"action":"create-story"}),
                );
            }
            "tray-sync" => {
                let _ = app.emit("desktop-action", serde_json::json!({"action":"sync"}));
            }
            "tray-quit" => app.exit(0),
            _ => {}
        });
    crate::tray_icon::configure_icon(tray, app).build(app)?;
    Ok(())
}
