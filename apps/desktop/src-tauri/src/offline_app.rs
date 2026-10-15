use crate::offline::{
    self, auth, collaboration, commands, drafts, files, native, recovery, settings, store::Store,
    transport::Transport, Runtime,
};
use std::sync::{Arc, Mutex};
use tauri::{
    ipc::CapabilityBuilder,
    menu::{Menu, MenuItem, Submenu},
    Emitter, Manager, WebviewUrl, WebviewWindowBuilder,
};
use tauri_plugin_deep_link::DeepLinkExt;
use tauri_plugin_opener::OpenerExt;

fn handle_link(app: &tauri::AppHandle, url: tauri::Url) {
    if url.scheme() != auth::CALLBACK_SCHEME {
        return;
    }
    if url.host_str() == Some("login") {
        let app = app.clone();
        tauri::async_runtime::spawn(async move {
            if let Err(error) = auth::callback(app.clone(), url).await {
                let _ = app.emit("desktop-error", error);
            }
        });
        return;
    }
    if url.host_str() != Some("workspace")
        || url.query().is_some()
        || url.fragment().is_some()
        || !url.username().is_empty()
        || url.password().is_some()
    {
        return;
    }
    let parts: Vec<_> = url.path().trim_matches('/').split('/').collect();
    if parts.len() != 3
        || !matches!(parts[1], "story" | "document")
        || uuid::Uuid::parse_str(parts[2]).is_err()
    {
        return;
    }
    if let Err(error) = offline::activation::activate(
        app,
        serde_json::json!({"workspaceId":parts[0],"kind":parts[1],"resourceId":parts[2]}),
    ) {
        let _ = app.emit("desktop-error", error);
    }
}

pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_single_instance::init(|app, args, _| {
            if let Some(window) = app.get_webview_window("main") {
                let _ = window.show();
                let _ = window.set_focus();
            }
            for arg in args {
                if let Ok(url) = tauri::Url::parse(&arg) {
                    handle_link(app, url);
                }
            }
        }))
        .plugin(tauri_plugin_deep_link::init())
        .plugin(
            tauri_plugin_opener::Builder::new()
                .open_js_links_on_click(false)
                .build(),
        )
        .plugin(tauri_plugin_window_state::Builder::default().build())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_updater::Builder::new().build())
        .plugin(tauri_plugin_autostart::Builder::new().build())
        .invoke_handler(tauri::generate_handler![
            commands::desktop_bootstrap,
            commands::desktop_select_workspace,
            commands::desktop_read_workspace,
            commands::desktop_save_operation,
            commands::desktop_resolve_operation,
            commands::desktop_logout,
            commands::desktop_open_external,
            auth::desktop_sign_in,
            offline::sync::desktop_sync,
            drafts::desktop_read_draft,
            drafts::desktop_save_draft,
            drafts::desktop_delete_draft,
            files::desktop_stage_attachment,
            native::desktop_settings,
            native::desktop_update_settings,
            native::desktop_check_update,
            native::desktop_install_update,
            collaboration::desktop_collaboration_session,
            recovery::desktop_export_document,
            recovery::desktop_export_local_work,
            drafts::desktop_pending_collaborations,
            drafts::desktop_ack_collaboration,
            settings::desktop_read_settings,
            settings::desktop_sync_settings,
            settings::desktop_open_workspace_page,
            native::desktop_confirm_exit,
            offline::activation::desktop_take_navigation,
            offline::download::desktop_refresh_resource,
            offline::media::desktop_read_media,
            offline::media::desktop_cache_media,
            offline::media::desktop_export_attachment,
            offline::notifications::desktop_mark_notification_read,
        ])
        .setup(|app| {
            let directory = app.path().app_local_data_dir()?;
            std::fs::create_dir_all(&directory)?;
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                std::fs::set_permissions(&directory, std::fs::Permissions::from_mode(0o700))?;
            }
            let store = Store::open(&directory.join("workspace.sqlite3"))?;
            let transport = Transport::new(store.account().ok())?;
            let runtime = Arc::new(Runtime {
                store,
                transport,
                sync: tokio::sync::Mutex::new(()),
                login: Mutex::new(None),
                allow_exit: std::sync::atomic::AtomicBool::new(false),
            });
            offline::media::initialize(&runtime.store)?;
            app.manage(Arc::clone(&runtime));
            offline::notifications::start(
                app.handle(),
                Arc::clone(&runtime),
                offline::activation::activate,
            )?;
            offline::background::start(app.handle(), Arc::clone(&runtime));
            let mut capability = CapabilityBuilder::new("bundled-desktop")
                .local(true)
                .webview("main")
                .permission("core:event:allow-listen")
                .permission("core:event:allow-unlisten")
                .permission("core:window:allow-start-dragging");
            for command in [
                "desktop-bootstrap",
                "desktop-select-workspace",
                "desktop-read-workspace",
                "desktop-save-operation",
                "desktop-resolve-operation",
                "desktop-logout",
                "desktop-open-external",
                "desktop-sign-in",
                "desktop-sync",
                "desktop-read-draft",
                "desktop-save-draft",
                "desktop-delete-draft",
                "desktop-stage-attachment",
                "desktop-settings",
                "desktop-update-settings",
                "desktop-check-update",
                "desktop-install-update",
                "desktop-collaboration-session",
                "desktop-export-document",
                "desktop-export-local-work",
                "desktop-pending-collaborations",
                "desktop-ack-collaboration",
                "desktop-read-settings",
                "desktop-sync-settings",
                "desktop-open-workspace-page",
                "desktop-confirm-exit",
                "desktop-take-navigation",
                "desktop-refresh-resource",
                "desktop-read-media",
                "desktop-cache-media",
                "desktop-export-attachment",
                "desktop-mark-notification-read",
            ] {
                capability = capability.permission(format!("allow-{command}"));
            }
            app.add_capability(capability)?;
            let handle = app.handle().clone();
            let builder =
                WebviewWindowBuilder::new(app, "main", WebviewUrl::App("index.html".into()))
                    .title("FortyOne")
                    .inner_size(1280.0, 800.0)
                    .min_inner_size(960.0, 640.0)
                    .disable_drag_drop_handler()
                    .on_navigation(move |url| {
                        let local = offline::is_local_url(url)
                            || (cfg!(debug_assertions)
                                && url.origin().ascii_serialization() == "http://localhost:5178");
                        if !local && crate::navigation::can_open_externally(url) {
                            let _ = handle.opener().open_url(url.as_str(), None::<&str>);
                        }
                        local
                    });
            #[cfg(target_os = "macos")]
            let builder = builder
                .title_bar_style(tauri::TitleBarStyle::Overlay)
                .hidden_title(true);
            let window = builder.build()?;
            #[cfg(target_os = "macos")]
            {
                crate::macos_chrome::install(&window)?;
            }
            offline::window::install(app.handle(), &window)?;
            let menu = Menu::default(app.handle())?;
            menu.append(&Submenu::with_items(
                app,
                "Workspace",
                true,
                &[
                    &MenuItem::with_id(
                        app,
                        "desktop-new",
                        "New Story…",
                        true,
                        Some("CmdOrCtrl+N"),
                    )?,
                    &MenuItem::with_id(
                        app,
                        "desktop-search",
                        "Search…",
                        true,
                        Some("CmdOrCtrl+K"),
                    )?,
                    &MenuItem::with_id(
                        app,
                        "desktop-sync",
                        "Sync Now",
                        true,
                        Some("CmdOrCtrl+Shift+R"),
                    )?,
                ],
            )?)?;
            app.set_menu(menu)?;
            app.on_menu_event(|app, event| {
                let action = match event.id().as_ref() {
                    "desktop-new" => "create-story",
                    "desktop-search" => "search",
                    "desktop-sync" => "sync",
                    _ => return,
                };
                let _ = app.emit("desktop-action", serde_json::json!({"action":action}));
            });
            let handle = app.handle().clone();
            app.deep_link().on_open_url(move |event| {
                for url in event.urls() {
                    handle_link(&handle, url);
                }
            });
            if let Some(urls) = app.deep_link().get_current()? {
                for url in urls {
                    handle_link(app.handle(), url);
                }
            }
            Ok(())
        })
        .build(crate::desktop_context())
        .expect("Could not start FortyOne desktop")
        .run(|app, event| {
            #[cfg(target_os = "macos")]
            if let tauri::RunEvent::Reopen { .. } = event {
                if let Some(window) = app.get_webview_window("main") {
                    let _ = window.show();
                    let _ = window.set_focus();
                }
            }
            if let tauri::RunEvent::ExitRequested { api, .. } = event {
                let runtime = app.state::<Arc<Runtime>>();
                if !runtime.allow_exit.load(std::sync::atomic::Ordering::SeqCst) {
                    api.prevent_exit();
                    let _ = app.emit("desktop-before-quit", ());
                    return;
                }
                if let Ok(db) = runtime.store.0.lock() {
                    let _ = db.execute_batch("PRAGMA wal_checkpoint(FULL)");
                };
            }
        });
}
