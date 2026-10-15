#[cfg(not(target_os = "macos"))]
use tauri::Manager;
use tauri::{tray::TrayIconBuilder, AppHandle, Wry};

pub fn configure_icon(builder: TrayIconBuilder<Wry>, app: &AppHandle) -> TrayIconBuilder<Wry> {
    #[cfg(target_os = "macos")]
    {
        let _ = app;
        builder
            .icon(tauri::include_image!("icons/tray-template.png"))
            .icon_as_template(true)
    }
    #[cfg(not(target_os = "macos"))]
    {
        match app.default_window_icon() {
            Some(icon) => builder.icon(icon.clone()),
            None => builder,
        }
    }
}
