(() => {
  if (window.top !== window) return;

  const installAuthChrome = () => {
    const style = document.createElement("style");
    // Keep the existing React header as a transparent drag region. Scoping to
    // mounted onboarding content also preserves workspace and settings chrome.
    style.textContent = `
      html[data-desktop-platform="macos"][data-desktop-fallback="true"]:has(.onboarding-layout) body {
        padding-top: 0;
      }
      html[data-desktop-platform="macos"][data-desktop-fallback="true"]:has(.onboarding-layout) .desktop-fallback-titlebar {
        height: 36px;
        background: transparent;
      }
      html[data-desktop-platform="macos"][data-desktop-fallback="true"]:has(.onboarding-layout) .desktop-fallback-titlebar > [data-desktop-history] {
        display: none;
      }
    `;
    document.head.appendChild(style);
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", installAuthChrome, {
      once: true,
    });
  } else {
    installAuthChrome();
  }

  // Preserve Next.js history metadata; the native webview owns the actual stack.
  for (const method of ["pushState", "replaceState"]) {
    const original = window.history[method];
    window.history[method] = function (...args) {
      const result = Reflect.apply(original, this, args);
      window.dispatchEvent(new Event("fortyone:desktop-history-change"));
      return result;
    };
  }

  Object.defineProperty(window, "__FORTYONE_DESKTOP__", {
    value: Object.freeze({
      version: 2,
      platform: "macos",
      vibrancy: true,
      syncChrome: (state) =>
        window.__TAURI_INTERNALS__.invoke("sync_desktop_chrome", { state }),
      navigationState: () =>
        window.__TAURI_INTERNALS__.invoke("desktop_navigation_state"),
      navigate: (direction) =>
        window.__TAURI_INTERNALS__.invoke("desktop_navigate", { direction }),
      windowAction: (action) =>
        window.__TAURI_INTERNALS__.invoke("desktop_window_action", { action }),
    }),
    writable: false,
    configurable: false,
  });
})();
