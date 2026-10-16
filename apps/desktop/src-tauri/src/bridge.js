(() => {
  if (window.top !== window) return;
  const platform = "__FORTYONE_DESKTOP_PLATFORM__";

  let closeFreeze = null;
  const restoreCloseInput = () => {
    if (!closeFreeze) return;
    closeFreeze.body.inert = closeFreeze.inert;
    closeFreeze = null;
  };
  window.addEventListener("fortyone:desktop-close-failed", restoreCloseInput);
  window.addEventListener("fortyone:native-before-close", async (event) => {
    const nonce = event.detail?.nonce;
    if (typeof nonce !== "string") return;
    if (!closeFreeze && document.body) {
      closeFreeze = { body: document.body, inert: document.body.inert };
      document.body.inert = true;
    }
    const pending = [];
    window.dispatchEvent(new CustomEvent("fortyone:desktop-before-close", {
      detail: { waitUntil: (promise) => pending.push(Promise.resolve(promise)) },
    }));
    try {
      await Promise.all(pending);
      await window.__TAURI_INTERNALS__.invoke("hosted_confirm_close", { nonce, allow: true });
      if (event.detail?.action === "hide") restoreCloseInput();
    } catch (error) {
      restoreCloseInput();
      const message = typeof error === "string" ? error : error?.message || "Could not save your local drafts";
      try {
        await window.__TAURI_INTERNALS__.invoke("hosted_confirm_close", { nonce, allow: false, error: String(message).slice(0, 500) });
      } catch {
        // The native timeout keeps the app open if its pending request expired.
      }
    }
  });

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
      platform,
      vibrancy: platform === "macos",
      exportRecovery: (request) => window.__TAURI_INTERNALS__.invoke("hosted_export_recovery", { request }),
      stageAttachment: async (request) => {
        const result = await window.__TAURI_INTERNALS__.invoke("hosted_stage_attachment", { request });
        window.dispatchEvent(new CustomEvent("fortyone:desktop-outbox-changed", { detail: { accountId: request.accountId, workspaceId: request.workspaceId } }));
        return result;
      },
      readAttachment: (request) => window.__TAURI_INTERNALS__.invoke("hosted_read_attachment", { request }),
      exportAttachment: (request) => window.__TAURI_INTERNALS__.invoke("hosted_export_attachment", { request }),
      pageRead: (request) => window.__TAURI_INTERNALS__.invoke("hosted_page_read", { request }),
      pageWrite: (request) => window.__TAURI_INTERNALS__.invoke("hosted_page_write", { request }),
      enqueueOperation: async (request) => {
        const result = await window.__TAURI_INTERNALS__.invoke("hosted_enqueue_operation", { request });
        window.dispatchEvent(new CustomEvent("fortyone:desktop-outbox-changed", { detail: { accountId: request.accountId, workspaceId: request.workspaceId } }));
        return result;
      },
      readOutbox: (request) => window.__TAURI_INTERNALS__.invoke("hosted_read_outbox", { request }),
      syncOutbox: (request) => window.__TAURI_INTERNALS__.invoke("hosted_sync_outbox", { request }),
      resolveOperation: async (request) => {
        const result = await window.__TAURI_INTERNALS__.invoke("hosted_resolve_operation", { request });
        window.dispatchEvent(new CustomEvent("fortyone:desktop-outbox-changed", { detail: { accountId: request.accountId, workspaceId: request.workspaceId } }));
        return result;
      },
      bindSession: (request) => window.__TAURI_INTERNALS__.invoke("hosted_bind_session", { request }),
      restoreSession: (request) => window.__TAURI_INTERNALS__.invoke("hosted_restore_session", { request }),
      cacheRead: (request) => window.__TAURI_INTERNALS__.invoke("hosted_cache_read", { request }),
      cacheWrite: (request) => window.__TAURI_INTERNALS__.invoke("hosted_cache_write", { request }),
      cacheClear: (request) => window.__TAURI_INTERNALS__.invoke("hosted_cache_clear", { request }),
      draftRead: (request) => window.__TAURI_INTERNALS__.invoke("hosted_draft_read", { request }),
      draftWrite: (request) => window.__TAURI_INTERNALS__.invoke("hosted_draft_write", { request }),
      draftDelete: (request) => window.__TAURI_INTERNALS__.invoke("hosted_draft_delete", { request }),
      getUpdateStatus: () => window.__TAURI_INTERNALS__.invoke("hosted_get_update_status"),
      checkForUpdate: () => window.__TAURI_INTERNALS__.invoke("hosted_check_update"),
      installUpdate: (request) => window.__TAURI_INTERNALS__.invoke("hosted_install_update", { request }),
      getPreferences: () => window.__TAURI_INTERNALS__.invoke("hosted_get_preferences"),
      updatePreferences: (preferences) => window.__TAURI_INTERNALS__.invoke("hosted_update_preferences", { preferences }),
      requestNotificationPermission: () => window.__TAURI_INTERNALS__.invoke("hosted_request_notification_permission"),
      deliverNotification: (notification) => window.__TAURI_INTERNALS__.invoke("hosted_deliver_notification", { notification }),
      setNotificationBadge: (state) => window.__TAURI_INTERNALS__.invoke("hosted_set_notification_badge", { state }),
      beginWorkspaceSwitch: async (destination) => {
        await window.__TAURI_INTERNALS__.invoke("begin_workspace_switch", { destination });
        window.dispatchEvent(new Event("fortyone:workspace-switch-change"));
      },
      workspaceReady: async (workspace) => {
        await window.__TAURI_INTERNALS__.invoke("workspace_ready", { workspace });
        window.dispatchEvent(new Event("fortyone:workspace-switch-change"));
      },
      cancelWorkspaceSwitch: async () => {
        await window.__TAURI_INTERNALS__.invoke("cancel_workspace_switch");
        window.dispatchEvent(new Event("fortyone:workspace-switch-change"));
      },
      syncMenu: (state) =>
        window.__TAURI_INTERNALS__.invoke("sync_desktop_menu", { state }),
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
