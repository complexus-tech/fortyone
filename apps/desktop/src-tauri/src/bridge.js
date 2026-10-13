(() => {
  if (window.top !== window) return;

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
