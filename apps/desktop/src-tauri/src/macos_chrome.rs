use objc2_app_kit::{
    NSAppearance, NSAppearanceCustomization, NSAppearanceNameAqua, NSAppearanceNameDarkAqua,
    NSColor, NSTitlebarSeparatorStyle, NSWindow, NSWindowButton, NSWindowStyleMask,
    NSWindowTitleVisibility,
};
use objc2_web_kit::WKWebView;
use tauri::WebviewWindow;

use crate::chrome::ChromeState;

/// Logical points; the shared desktop header uses the same height in CSS.
pub const TITLEBAR_HEIGHT: f64 = 44.0;
pub const TRAFFIC_LIGHT_LEFT: f64 = 20.0;

async fn with_native_webview<T: Send + 'static>(
    window: &WebviewWindow,
    operation: impl FnOnce(&WKWebView) -> T + Send + 'static,
) -> Result<T, String> {
    let (sender, mut receiver) = tauri::async_runtime::channel(1);
    window
        .with_webview(move |webview| {
            // SAFETY: Tauri exposes this webview's live WKWebView and runs the
            // callback on the UI thread. Only the operation's owned result can
            // leave the callback; the borrowed native view remains on it.
            let result = unsafe { webview.inner().cast::<WKWebView>().as_ref() }
                .map(operation)
                .ok_or_else(|| "Could not access the native macOS webview".to_string());

            // The channel has one slot and one sender. Failure only means that
            // the caller cancelled its request; never block the UI thread.
            let _ = sender.try_send(result);
        })
        .map_err(|error| error.to_string())?;

    receiver
        .recv()
        .await
        .ok_or_else(|| "The native macOS webview closed before responding".to_string())?
}

/// Reads WebKit's actual history, including same-document pushState entries.
pub async fn navigation_state(window: &WebviewWindow) -> Result<(bool, bool), String> {
    with_native_webview(window, |webview| {
        // SAFETY: with_native_webview supplies a live WKWebView on the UI
        // thread. Both selectors are supported by every target macOS version.
        unsafe { (webview.canGoBack(), webview.canGoForward()) }
    })
    .await
}

/// Initiates navigation on the UI thread; navigating beyond history is a no-op.
pub async fn navigate(window: &WebviewWindow, back: bool) -> Result<(), String> {
    with_native_webview(window, move |webview| {
        // SAFETY: with_native_webview guarantees the view's lifetime and UI
        // thread. Navigation objects are released here and never sent across
        // threads; WebKit owns the in-flight navigation independently.
        unsafe {
            if back && webview.canGoBack() {
                let _ = webview.goBack();
            } else if !back && webview.canGoForward() {
                let _ = webview.goForward();
            }
        }
    })
    .await
}

fn with_native_window(
    window: &WebviewWindow,
    operation: impl FnOnce(&NSWindow) + Send + 'static,
) -> tauri::Result<()> {
    window.with_webview(move |webview| {
        // SAFETY: Tauri provides the live NSWindow associated with this webview
        // and runs this callback on the UI thread. The reference never escapes
        // the callback, and no AppKit objects are moved between threads.
        let Some(native_window) = (unsafe { webview.ns_window().cast::<NSWindow>().as_ref() })
        else {
            eprintln!("Could not access the native macOS window");
            return;
        };
        operation(native_window);
    })
}

fn position_traffic_lights(window: &NSWindow) {
    // Fullscreen uses AppKit's separate, auto-hiding title bar. Reposition the
    // regular controls when the window returns to its normal layout instead.
    if window.styleMask().contains(NSWindowStyleMask::FullScreen) {
        return;
    }
    let (Some(close), Some(minimize)) = (
        window.standardWindowButton(NSWindowButton::CloseButton),
        window.standardWindowButton(NSWindowButton::MiniaturizeButton),
    ) else {
        return;
    };
    // SAFETY: The window owns these live AppKit views, and this entire layout
    // pass runs synchronously on the UI thread via with_native_window.
    let Some(container) = (unsafe { close.superview().and_then(|view| view.superview()) }) else {
        return;
    };

    // AppKit restores this container's default height after appearance and
    // window-size changes. Place it after those changes instead of relying on
    // Wry's initial inset, which is otherwise only reapplied when its view draws.
    let spacing = minimize.frame().origin.x - close.frame().origin.x;
    let window_height = window.frame().size.height;
    let mut container_frame = container.frame();
    container_frame.size.height = TITLEBAR_HEIGHT;
    container_frame.origin.y = window_height - TITLEBAR_HEIGHT;
    if container.frame() != container_frame {
        container.setFrame(container_frame);
    }

    let buttons = [
        Some(close),
        Some(minimize),
        window.standardWindowButton(NSWindowButton::ZoomButton),
    ];
    for (index, button) in buttons.into_iter().enumerate() {
        let Some(button) = button else { continue };
        // SAFETY: The retained standard button belongs to the same live window;
        // its parent is accessed and used only during this UI-thread callback.
        let Some(parent) = (unsafe { button.superview() }) else {
            continue;
        };
        // Convert through window coordinates so AppKit's internal button
        // origins and flipped views cannot change our visible center line.
        let mut target = button.convertRect_toView(button.bounds(), None);
        target.origin.x = TRAFFIC_LIGHT_LEFT + index as f64 * spacing;
        target.origin.y = window_height - (TITLEBAR_HEIGHT + target.size.height) / 2.0;
        let origin = parent.convertRect_fromView(target, None).origin;
        if button.frame().origin != origin {
            button.setFrameOrigin(origin);
        }
    }
}

/// Reapplies the shared header geometry after native resize or display changes.
/// Call from a worker when handling a window event, so this UI callback runs
/// after AppKit finishes the event's own title-bar layout.
pub fn refresh_layout(window: &WebviewWindow) -> tauri::Result<()> {
    with_native_window(window, position_traffic_lights)
}

/// Keeps native window controls above the app's single shared header. Tauri's
/// overlay title-bar configuration owns the full-size webview layout.
pub fn install(window: &WebviewWindow) -> tauri::Result<()> {
    with_native_window(window, |native_window| {
        native_window.setTitleVisibility(NSWindowTitleVisibility::Hidden);
        native_window.setTitlebarAppearsTransparent(true);
        native_window.setTitlebarSeparatorStyle(NSTitlebarSeparatorStyle::None);
        position_traffic_lights(native_window);
    })
}

/// Matches the native frame to the web theme without creating another toolbar.
pub fn update(window: &WebviewWindow, state: ChromeState) -> tauri::Result<()> {
    with_native_window(window, move |native_window| {
        let [red, green, blue] = state.background.map(|channel| f64::from(channel) / 255.0);
        native_window.setBackgroundColor(Some(&NSColor::colorWithSRGBRed_green_blue_alpha(
            red, green, blue, 1.0,
        )));
        let appearance = if state.system_theme {
            None
        } else {
            // SAFETY: These immutable AppKit appearance names are available on
            // every supported macOS version and accessed on the main thread.
            let name = unsafe {
                if state.dark {
                    NSAppearanceNameDarkAqua
                } else {
                    NSAppearanceNameAqua
                }
            };
            NSAppearance::appearanceNamed(name)
        };
        native_window.setAppearance(appearance.as_deref());
        position_traffic_lights(native_window);
    })
}
