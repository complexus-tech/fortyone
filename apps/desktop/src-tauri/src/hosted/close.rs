//! Closing is a renderer durability boundary, not a request to synchronize the network.
use std::{sync::Mutex, time::Duration};
use tauri::{AppHandle, Manager, WebviewWindow};
use tauri_plugin_dialog::{DialogExt, MessageDialogButtons, MessageDialogResult};

#[derive(Clone, Copy, PartialEq, Debug)]
pub(super) enum Action {
    Hide,
    Exit,
    Update,
}

struct Pending {
    nonce: String,
    origin: String,
    action: Action,
}

#[derive(Default)]
struct Gate {
    pending: Option<Pending>,
    failed: bool,
    dialog: Option<String>,
    allow_exit: bool,
    update_waiter: Option<tokio::sync::oneshot::Sender<Result<(), String>>>,
}

#[derive(Default)]
pub(super) struct CloseState(Mutex<Gate>);

impl Gate {
    fn confirm(&mut self, nonce: &str, origin: &str) -> Result<Action, String> {
        let pending = self
            .pending
            .as_ref()
            .ok_or("There is no pending close request")?;
        if pending.nonce != nonce || pending.origin != origin {
            return Err("The close request belongs to another page".into());
        }
        let action = pending.action;
        self.pending = None;
        Ok(action)
    }
}

fn finish(app: &AppHandle, action: Action) {
    // app.exit() may finish without another ExitRequested callback. Save before
    // either native transition, including explicit close-without-saving recovery.
    if let Some(window) = app.get_webview_window("main") {
        if let Err(error) = crate::session::save_window(&window) {
            if let Ok(mut gate) = app.state::<CloseState>().0.lock() {
                gate.failed = true;
            }
            failure(
                &window,
                &format!("Could not remember the open workspace: {error}"),
            );
            return;
        }
    }
    if action == Action::Exit {
        if let Ok(mut gate) = app.state::<CloseState>().0.lock() {
            gate.allow_exit = true;
        }
        app.exit(0);
    } else if action == Action::Hide {
        if let Some(window) = app.get_webview_window("main") {
            let _ = window.hide();
        }
    }
}

pub(super) fn failure(window: &WebviewWindow, message: &str) {
    let message: String = message.chars().take(1024).collect();
    super::preferences::show(window.app_handle());
    let detail = serde_json::json!({"message":message});
    let _ = window.eval(format!("window.dispatchEvent(new CustomEvent('fortyone:desktop-close-failed',{{detail:{detail}}}))"));
}

fn recover(window: &WebviewWindow, action: Action, origin: String, nonce: String) {
    super::preferences::show(window.app_handle());
    let app = window.app_handle().clone();
    window.dialog()
        .message("FortyOne could not save every change on this computer. You can retry saving or stay open to recover your work. Closing without saving may lose the latest changes.")
        .title("Some changes are not saved")
        .buttons(MessageDialogButtons::YesNoCancelCustom("Retry saving".into(),"Close without saving".into(),"Stay open".into()))
        .parent(window)
        .show_with_result(move |choice| {
            let Some(window) = app.get_webview_window("main") else { return; };
            let state = app.state::<CloseState>();
            let Ok(mut gate) = state.0.lock() else { return; };
            if gate.dialog.as_deref() != Some(nonce.as_str()) { return; }
            gate.dialog = None;
            if window.url().ok().map(|url|url.origin().ascii_serialization()).as_deref() != Some(origin.as_str()) { return; }
            let retry = matches!(&choice, MessageDialogResult::Yes) || matches!(&choice,MessageDialogResult::Custom(label) if label == "Retry saving");
            let discard = matches!(&choice, MessageDialogResult::No) || matches!(&choice,MessageDialogResult::Custom(label) if label == "Close without saving");
            if retry || discard { gate.failed = false; }
            drop(gate);
            if retry { request(&window, action); }
            else if discard { finish(&app, action); }
        });
}

pub(super) fn request(window: &WebviewWindow, action: Action) {
    if super::update::in_progress() {
        failure(
            window,
            "An update is in progress. Wait for it to finish before closing FortyOne.",
        );
        return;
    }
    let app = window.app_handle();
    let state = app.state::<CloseState>();
    let Ok(mut gate) = state.0.lock() else {
        return;
    };
    if let Some(pending) = gate.pending.as_mut() {
        // An explicit Quit upgrades an in-flight window-close request.
        if action == Action::Exit && pending.action != Action::Update {
            pending.action = action;
        }
        return;
    }
    if gate.dialog.is_some() {
        return;
    }
    let origin = match window.url() {
        Ok(url) => url.origin().ascii_serialization(),
        Err(_) => {
            drop(gate);
            failure(
                window,
                "FortyOne could not verify the open page. Keep the window open and try again.",
            );
            return;
        }
    };
    let nonce = uuid::Uuid::new_v4().to_string();
    if gate.failed {
        gate.dialog = Some(nonce.clone());
        drop(gate);
        recover(window, action, origin, nonce);
        return;
    }
    gate.pending = Some(Pending {
        nonce: nonce.clone(),
        origin,
        action,
    });
    drop(gate);
    let detail = serde_json::json!({"nonce":nonce,"action":if action == Action::Hide {"hide"} else {"exit"}});
    if window.eval(format!("window.dispatchEvent(new CustomEvent('fortyone:native-before-close',{{detail:{detail}}}))")).is_err() {
        if let Ok(mut gate) = state.0.lock() { gate.pending = None; gate.failed = true; }
        failure(window,"FortyOne could not ask the page to save. Keep the window open and try again.");
        return;
    }
    let window = window.clone();
    tauri::async_runtime::spawn(async move {
        tokio::time::sleep(Duration::from_secs(12)).await;
        let state = window.state::<CloseState>();
        let timed_out = if let Ok(mut gate) = state.0.lock() {
            if gate
                .pending
                .as_ref()
                .is_some_and(|pending| pending.nonce == nonce)
            {
                gate.pending = None;
                gate.failed = true;
                true
            } else {
                false
            }
        } else {
            false
        };
        if timed_out {
            failure(
                &window,
                "Saving took too long. Your window stayed open. Retry saving before closing.",
            );
        }
    });
}

#[tauri::command]
pub fn hosted_confirm_close(
    window: WebviewWindow,
    nonce: String,
    allow: bool,
    error: Option<String>,
) -> Result<(), String> {
    super::authorize_window(&window)?;
    let origin = window
        .url()
        .map_err(|error| error.to_string())?
        .origin()
        .ascii_serialization();
    let state = window.state::<CloseState>();
    let (action, update_waiter) = {
        let mut gate = state.0.lock().map_err(|error| error.to_string())?;
        let action = gate.confirm(&nonce, &origin)?;
        gate.failed = !allow;
        (action, gate.update_waiter.take())
    };
    if action == Action::Update {
        let result = if allow {
            Ok(())
        } else {
            Err(error.unwrap_or_else(|| {
                "Some changes could not be saved. The update was stopped.".into()
            }))
        };
        if let Err(message) = &result {
            failure(&window, message);
        }
        if let Some(waiter) = update_waiter {
            let _ = waiter.send(result);
        }
        return Ok(());
    }
    if allow {
        finish(window.app_handle(), action);
    } else {
        failure(
            &window,
            error.as_deref().unwrap_or(
                "Some changes could not be saved. Your window stayed open so you can retry.",
            ),
        );
    }
    Ok(())
}

pub(super) fn allow_installer_exit(app: &AppHandle, allow: bool) -> Result<(), String> {
    if allow {
        if let Some(window) = app.get_webview_window("main") {
            crate::session::save_window(&window)?;
        }
    }
    let state = app.state::<CloseState>();
    state
        .0
        .lock()
        .map_err(|error| error.to_string())?
        .allow_exit = allow;
    Ok(())
}

/// Downloading may take time; flush only immediately before the installer runs.
pub(super) async fn flush_for_update(window: &WebviewWindow) -> Result<(), String> {
    let state = window.state::<CloseState>();
    let nonce = uuid::Uuid::new_v4().to_string();
    let (sender, receiver) = tokio::sync::oneshot::channel();
    {
        let mut gate = state.0.lock().map_err(|error| error.to_string())?;
        if gate.pending.is_some() || gate.dialog.is_some() {
            return Err("Finish the pending close request before installing an update".into());
        }
        gate.pending = Some(Pending {
            nonce: nonce.clone(),
            origin: window
                .url()
                .map_err(|error| error.to_string())?
                .origin()
                .ascii_serialization(),
            action: Action::Update,
        });
        gate.update_waiter = Some(sender);
    }
    let detail = serde_json::json!({"nonce":nonce,"action":"update"});
    let dispatched = window.eval(format!(
        "window.dispatchEvent(new CustomEvent('fortyone:native-before-close',{{detail:{detail}}}))"
    ));
    let result = if dispatched.is_ok() {
        match tokio::time::timeout(Duration::from_secs(12), receiver).await {
            Ok(Ok(result)) => result,
            _ => Err(
                "Saving took too long. The update was stopped and your window stayed open.".into(),
            ),
        }
    } else {
        Err("The page could not save its changes. The update was stopped.".into())
    };
    if let Ok(mut gate) = state.0.lock() {
        if gate
            .pending
            .as_ref()
            .is_some_and(|pending| pending.nonce == nonce)
        {
            gate.pending = None;
            gate.update_waiter = None;
        }
    }
    if let Err(message) = &result {
        failure(window, message);
    }
    result
}

/// True means the caller must prevent this ExitRequested event.
pub fn prevent_exit(app: &AppHandle) -> bool {
    let Some(state) = app.try_state::<CloseState>() else {
        return false;
    };
    if state.0.lock().is_ok_and(|gate| gate.allow_exit) {
        return false;
    }
    let Some(window) = app.get_webview_window("main") else {
        return false;
    };
    request(&window, Action::Exit);
    true
}

#[cfg(test)]
mod tests {
    use super::*;
    fn gate() -> Gate {
        Gate {
            pending: Some(Pending {
                nonce: "nonce".into(),
                origin: "https://team.fortyone.app".into(),
                action: Action::Hide,
            }),
            ..Gate::default()
        }
    }
    #[test]
    fn only_the_requested_page_can_confirm_once() {
        let mut gate = gate();
        assert!(gate.confirm("wrong", "https://team.fortyone.app").is_err());
        assert!(gate.confirm("nonce", "https://other.fortyone.app").is_err());
        assert!(gate.pending.is_some());
        assert_eq!(
            gate.confirm("nonce", "https://team.fortyone.app").unwrap(),
            Action::Hide
        );
        assert!(gate.confirm("nonce", "https://team.fortyone.app").is_err());
    }
    #[test]
    fn expired_requests_cannot_authorize_a_later_close() {
        let mut gate = gate();
        gate.pending.as_mut().unwrap().nonce = "new".into();
        assert!(gate.confirm("nonce", "https://team.fortyone.app").is_err());
        assert!(gate.pending.is_some());
        assert!(!gate.allow_exit);
    }
}
