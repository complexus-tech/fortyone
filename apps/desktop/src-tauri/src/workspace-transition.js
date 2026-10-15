(() => {
  if (window.top !== window) return;
  const ROOT_ID = "fortyone-native-workspace-switch";
  const STYLE_ID = `${ROOT_ID}-style`;
  let timer;
  let request = 0;

  const invoke = (command) => window.__TAURI_INTERNALS__.invoke(command);
  const remove = () => document.getElementById(ROOT_ID)?.remove();
  const installStyle = () => {
    if (document.getElementById(STYLE_ID) || !document.head) return;
    const style = document.createElement("style");
    style.id = STYLE_ID;
    style.textContent = `
      #${ROOT_ID} { position:fixed; bottom:24px; left:50%; transform:translateX(-50%); z-index:2147483646; max-width:calc(100vw - 48px); width:max-content; box-sizing:border-box; border:1px solid var(--color-border,#d4d4d4); border-radius:12px; background:var(--color-surface-elevated,#fff); color:var(--color-foreground,#171717); padding:14px 18px; box-shadow:0 6px 24px #0002; font:14px/1.5 system-ui,sans-serif; }
      #${ROOT_ID} .workspace-switch-heading { display:flex; align-items:center; gap:10px; font-weight:500; overflow-wrap:anywhere; }
      #${ROOT_ID} .workspace-switch-spinner { flex:none; width:14px; height:14px; border:2px solid currentColor; border-right-color:transparent; border-radius:50%; animation:fortyone-workspace-spin .7s linear infinite; }
      #${ROOT_ID} .workspace-switch-description { margin:6px 0 0; opacity:.75; }
      #${ROOT_ID} .workspace-switch-actions { display:flex; gap:8px; margin-top:12px; }
      #${ROOT_ID} button { padding:6px 12px; border:1px solid var(--color-border,#d4d4d4); background:transparent; color:inherit; border-radius:7px; font:inherit; cursor:pointer; }
      #${ROOT_ID} button:focus-visible { outline:2px solid currentColor; outline-offset:2px; }
      #${ROOT_ID} button:disabled { opacity:.5; cursor:progress; }
      @keyframes fortyone-workspace-spin { to { transform:rotate(360deg); } }
      @media(prefers-reduced-motion:reduce) { #${ROOT_ID} .workspace-switch-spinner { animation:none; } }
    `;
    document.head.appendChild(style);
  };

  const actionButton = (label, command, root) => {
    const button = document.createElement("button");
    button.type = "button";
    button.textContent = label;
    button.addEventListener("click", async () => {
      for (const control of root.querySelectorAll("button"))
        control.disabled = true;
      try {
        await invoke(command);
        await refresh();
      } catch (error) {
        console.error("Could not recover the desktop workspace switch", error);
        const message = root.querySelector(".workspace-switch-description");
        if (message)
          message.textContent =
            "The workspace could not be opened. Check your connection and try again.";
        for (const control of root.querySelectorAll("button"))
          control.disabled = false;
      }
    });
    return button;
  };

  const render = (state) => {
    if (!document.body) return;
    if (document.querySelector("[data-workspace-switch-progress]")) {
      remove();
      return;
    }
    installStyle();
    const current = document.getElementById(ROOT_ID);
    const identity = `${state.id}:${state.startedAt}:${state.failed}`;
    if (current?.dataset.transition === identity) return;
    remove();
    const root = document.createElement("div");
    root.id = ROOT_ID;
    root.dataset.transition = identity;
    root.setAttribute("role", state.failed ? "alert" : "status");
    root.setAttribute("aria-live", "polite");
    const heading = document.createElement("div");
    heading.className = "workspace-switch-heading";
    if (!state.failed) {
      const spinner = document.createElement("span");
      spinner.className = "workspace-switch-spinner";
      spinner.setAttribute("aria-hidden", "true");
      heading.appendChild(spinner);
    }
    const label = document.createElement("span");
    label.textContent = state.failed
      ? `${state.name} is taking longer to open.`
      : `Opening ${state.name}…`;
    heading.appendChild(label);
    root.appendChild(heading);
    if (state.failed) {
      const description = document.createElement("p");
      description.className = "workspace-switch-description";
      description.textContent =
        "Check your connection, then retry or return to your previous workspace.";
      root.appendChild(description);
      const actions = document.createElement("div");
      actions.className = "workspace-switch-actions";
      actions.append(
        actionButton("Retry", "retry_workspace_switch", root),
        actionButton(
          "Return to previous workspace",
          "cancel_workspace_switch",
          root,
        ),
      );
      root.appendChild(actions);
    }
    document.body.appendChild(root);
  };

  const refresh = async () => {
    const latestRequest = ++request;
    clearTimeout(timer);
    try {
      const state = await invoke("get_workspace_switch");
      if (latestRequest !== request) return;
      if (
        !state ||
        typeof state.id !== "string" ||
        typeof state.name !== "string"
      ) {
        remove();
        return;
      }
      render(state);
      // Poll only during an active transition; native state survives each document.
      timer = setTimeout(refresh, 500);
    } catch (error) {
      if (latestRequest !== request) return;
      console.error("Could not restore desktop workspace progress", error);
      remove();
    }
  };

  window.addEventListener("fortyone:workspace-switch-change", refresh);
  window.addEventListener("pageshow", refresh);
  window.addEventListener("pagehide", () => {
    ++request;
    clearTimeout(timer);
  });
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", refresh, { once: true });
  }
  void refresh();
})();
