export const FEEDBACK_WIDGET_LOADER_SOURCE = String.raw`(function (window, document) {
  "use strict";

  var currentScript = document.currentScript;
  var scriptOrigin = currentScript && currentScript.src
    ? new URL(currentScript.src, window.location.href).origin
    : window.location.origin;
  var previous = window.FortyOneFeedback;
  var active = null;
  var pendingIdentityAssertion = null;
  var hasPendingIdentityCommand = false;
  var identityCommandRevision = 0;
  var CHANNEL = "fortyone:feedback-widget";
  var VERSION = 1;

  function oneOf(value, values, fallback) {
    return values.indexOf(value) === -1 ? fallback : value;
  }

  function readOptions(script) {
    var data = script ? script.dataset : {};
    return {
      defaultTab: data.defaultTab,
      keyId: data.keyId,
      label: data.label,
      mode: data.mode,
      portalSlug: data.portal,
      position: data.position,
      target: data.target,
      theme: data.theme,
      trigger: data.trigger
    };
  }

  function normalizeOptions(options) {
    var portalSlug = String(options.portalSlug || options.portal || "").trim().toLowerCase();
    if (!/^(?=.{3,255}$)[a-z0-9](?:[a-z0-9-]*[a-z0-9])$/.test(portalSlug)) {
      throw new Error("FortyOne Feedback requires a valid portalSlug");
    }
    return {
      defaultTab: oneOf(options.defaultTab, ["home", "feedback", "roadmap", "updates"], "home"),
      keyId: options.keyId ? String(options.keyId) : "",
      label: String(options.label || "Share feedback"),
      mode: oneOf(options.mode, ["bubble", "custom", "inline"], "bubble"),
      portalSlug: portalSlug,
      position: oneOf(options.position, ["bottom-left", "bottom-right"], "bottom-right"),
      target: options.target ? String(options.target) : "#fortyone-feedback",
      theme: oneOf(options.theme, ["auto", "light", "dark"], "auto"),
      trigger: options.trigger ? String(options.trigger) : "[data-fortyone-feedback]"
    };
  }

  function makeInstanceId() {
    if (window.crypto && typeof window.crypto.randomUUID === "function") {
      return window.crypto.randomUUID();
    }
    return "fo-" + Date.now().toString(36) + "-" + Math.random().toString(36).slice(2);
  }

  function setIdentityCommand(assertion) {
    if (assertion === null || assertion === undefined || assertion === "") {
      pendingIdentityAssertion = null;
    } else {
      if (typeof assertion !== "string" || assertion.length > 16384) {
        throw new Error("FortyOne Feedback identify requires a signed assertion string");
      }
      pendingIdentityAssertion = assertion;
    }
    hasPendingIdentityCommand = true;
    identityCommandRevision += 1;
  }

  function createWidget(rawOptions) {
    var options = normalizeOptions(rawOptions || {});
    var instanceId = makeInstanceId();
    var iframe = null;
    var launcher = null;
    var open = false;
    var frameReady = false;
    var lastSentIdentityCommandRevision = 0;
    var destroyed = false;
    var restoreOverflow = null;
    var themeMedia = options.theme === "auto" && typeof window.matchMedia === "function"
      ? window.matchMedia("(prefers-color-scheme: dark)")
      : null;
    var target = options.mode === "inline" ? document.querySelector(options.target) : document.body;
    if (!target) {
      throw new Error("FortyOne Feedback could not find the inline target " + options.target);
    }

    var host = document.createElement("div");
    host.setAttribute("data-fortyone-feedback-root", "");
    var shadow = host.attachShadow({ mode: "open" });
    var style = document.createElement("style");
    style.textContent = [
      ":host{all:initial;color-scheme:light dark}",
      ".root{--widget-background:#fff;--widget-border:rgba(127,127,127,.26);color-scheme:light;font-family:-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif;position:relative;z-index:2147483000}",
      ".root[data-theme=dark]{--widget-background:oklch(0.1821 0.0139 94);--widget-border:#292824;color-scheme:dark}",
      ".launcher{position:fixed;bottom:20px;width:46px;height:46px;border:0;border-radius:999px;display:grid;place-items:center;background:#171717;color:#fff;box-shadow:0 12px 32px rgba(0,0,0,.22),inset 0 0 0 1px rgba(255,255,255,.14);cursor:pointer;transition:transform .2s ease,box-shadow .2s ease;z-index:2}",
      ".launcher:hover{transform:translateY(-2px);box-shadow:0 16px 38px rgba(0,0,0,.28),inset 0 0 0 1px rgba(255,255,255,.18)}",
      ".launcher:focus-visible{outline:3px solid rgba(59,130,246,.55);outline-offset:3px}",
      ".launcher svg{width:18px;height:18px;transition:transform .2s ease}",
      ".root[data-open=true] .launcher svg{transform:scale(.9) rotate(3deg)}",
      ".right .launcher{right:20px}.left .launcher{left:20px}",
      ".panel{position:fixed;bottom:78px;width:min(408px,calc(100vw - 32px));height:min(680px,calc(100dvh - 98px));overflow:hidden;border:1px solid var(--widget-border);border-radius:.825rem;background:var(--widget-background);box-shadow:0 30px 80px rgba(0,0,0,.25),0 6px 22px rgba(0,0,0,.1);opacity:0;visibility:hidden;transform:translateY(12px) scale(.985);transform-origin:bottom;transition:opacity .18s ease,transform .22s cubic-bezier(.2,.8,.2,1),visibility 0s linear .22s}",
      ".right .panel{right:20px}.left .panel{left:20px}",
      ".root[data-open=true] .panel{opacity:1;visibility:visible;transform:none;transition-delay:0s}",
      ".panel iframe{display:block;width:100%;height:100%;border:0;background:var(--widget-background);color-scheme:inherit}",
      "@supports(corner-shape:squircle){.panel{border-radius:2.2rem;corner-shape:squircle}}",
      ".inline .panel{position:relative;inset:auto;width:100%;height:640px;min-height:360px;opacity:1;visibility:visible;transform:none;box-shadow:0 8px 30px rgba(0,0,0,.1)}",
      "@media(max-width:640px){.popup .panel{inset:0;width:100vw;height:100dvh;max-height:none;border:0;border-radius:0;transform:translateY(12px)}.popup.root[data-open=true] .panel{transform:none}.launcher{bottom:14px}.right .launcher{right:14px}.left .launcher{left:14px}}",
      "@media(prefers-reduced-motion:reduce){.launcher,.panel,.launcher svg{transition:none}}"
    ].join("");

    var root = document.createElement("div");
    root.className = "root " + (options.mode === "inline" ? "inline" : "popup") + " " + (options.position === "bottom-left" ? "left" : "right");
    root.dataset.open = options.mode === "inline" ? "true" : "false";
    function syncShellTheme() {
      root.dataset.theme = options.theme === "dark" || (options.theme === "auto" && themeMedia && themeMedia.matches)
        ? "dark"
        : "light";
    }
    syncShellTheme();
    if (themeMedia && typeof themeMedia.addEventListener === "function") {
      themeMedia.addEventListener("change", syncShellTheme);
    }
    var panel = document.createElement("div");
    panel.className = "panel";
    panel.id = "fortyone-feedback-panel-" + instanceId;
    root.appendChild(panel);

    if (options.mode === "bubble") {
      launcher = document.createElement("button");
      launcher.className = "launcher";
      launcher.type = "button";
      launcher.setAttribute("aria-controls", panel.id);
      launcher.setAttribute("aria-expanded", "false");
      launcher.setAttribute("aria-label", options.label);
      launcher.innerHTML = '<svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M8.5 7.5C8.5 6.48748 9.39543 5.5 10.5 5.5C11.6046 5.5 12.5 6.32081 12.5 7.33333C12.5 7.69831 12.3837 8.03837 12.1831 8.32406C11.5854 9.17553 10.5 9.98748 10.5 11" stroke-linecap="round"/><path d="M10.5 13.5H10.509" stroke-linecap="round" stroke-linejoin="round"/><path d="M8 19.5C9.05038 20.3697 10.3145 20.9238 11.7635 21.0188C12.9052 21.0937 14.0971 21.0936 15.2365 21.0188C15.6288 20.9931 16.0565 20.9007 16.4248 20.751C16.8345 20.5845 17.0395 20.5012 17.1437 20.5138C17.2478 20.5264 17.3989 20.6364 17.7011 20.8563C18.2339 21.244 18.9051 21.5225 19.9005 21.4986C20.4038 21.4865 20.6555 21.4804 20.7681 21.2909C20.8808 21.1013 20.7405 20.8389 20.4598 20.3141C20.0706 19.5862 19.824 18.7529 20.1977 18.0852C20.8413 17.1315 21.3879 16.0021 21.4678 14.7823C21.5107 14.1269 21.5107 13.4481 21.4678 12.7927C21.4146 11.9799 21.2173 11.2073 20.9012 10.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M12.2365 17.0188C15.5829 16.7993 18.2485 14.1315 18.4678 10.7823C18.5107 10.1269 18.5107 9.4481 18.4678 8.79268C18.2485 5.44345 15.5829 2.77563 12.2365 2.55611C11.0948 2.48122 9.90285 2.48137 8.76352 2.55611C5.41711 2.77563 2.75153 5.44345 2.53219 8.79268C2.48927 9.4481 2.48927 10.1269 2.53219 10.7823C2.61208 12.0021 3.15875 13.1315 3.80233 14.0852C4.17601 14.7529 3.92939 15.5862 3.54017 16.3141C3.25953 16.8389 3.11921 17.1013 3.23187 17.2909C3.34454 17.4804 3.59621 17.4865 4.09954 17.4986C5.09493 17.5225 5.76615 17.244 6.29894 16.8563C6.60112 16.6364 6.75221 16.5264 6.85635 16.5138C6.96048 16.5012 7.16541 16.5845 7.57521 16.751C7.94352 16.9007 8.37117 16.9931 8.76352 17.0188C9.90285 17.0936 11.0948 17.0937 12.2365 17.0188Z" stroke-linejoin="round"/></svg>';
      launcher.addEventListener("click", function () {
        open ? closeWidget() : openWidget();
      });
      root.appendChild(launcher);
    }

    shadow.appendChild(style);
    shadow.appendChild(root);
    target.appendChild(host);

    function send(event, payload) {
      if (!iframe || !iframe.contentWindow) return;
      iframe.contentWindow.postMessage({
        channel: CHANNEL,
        event: event,
        instanceId: instanceId,
        payload: payload || {},
        version: VERSION
      }, scriptOrigin);
    }

    function ensureFrame() {
      if (iframe) return iframe;
      iframe = document.createElement("iframe");
      var params = new URLSearchParams({
        instance: instanceId,
        mode: options.mode,
        parentOrigin: window.location.origin,
        tab: options.defaultTab,
        theme: options.theme
      });
      iframe.src = scriptOrigin + "/embed/feedback/" + encodeURIComponent(options.portalSlug) + "?" + params.toString();
      iframe.title = options.label;
      iframe.loading = options.mode === "inline" ? "eager" : "lazy";
      iframe.allow = "clipboard-write";
      iframe.referrerPolicy = "strict-origin-when-cross-origin";
      panel.appendChild(iframe);
      return iframe;
    }

    function sendIdentityCommand() {
      if (!frameReady || !hasPendingIdentityCommand) return;
      if (lastSentIdentityCommandRevision === identityCommandRevision && pendingIdentityAssertion !== null) return;
      lastSentIdentityCommandRevision = identityCommandRevision;
      var requestId = String(identityCommandRevision);
      if (pendingIdentityAssertion === null) {
        send("host-identity-clear", { requestId: requestId });
        return;
      }
      send("host-identify", { assertion: pendingIdentityAssertion, requestId: requestId });
    }

    function identify(assertion) {
      if (destroyed) return;
      setIdentityCommand(assertion);
      sendIdentityCommand();
    }

    function lockMobileScroll() {
      if (!window.matchMedia("(max-width: 640px)").matches || restoreOverflow !== null) return;
      restoreOverflow = document.documentElement.style.overflow;
      document.documentElement.style.overflow = "hidden";
    }

    function unlockMobileScroll() {
      if (restoreOverflow === null) return;
      document.documentElement.style.overflow = restoreOverflow;
      restoreOverflow = null;
    }

    function openWidget() {
      if (destroyed || options.mode === "inline") return;
      ensureFrame();
      open = true;
      root.dataset.open = "true";
      if (launcher) launcher.setAttribute("aria-expanded", "true");
      lockMobileScroll();
      send("host-open");
    }

    function closeWidget() {
      if (destroyed || options.mode === "inline") return;
      open = false;
      root.dataset.open = "false";
      if (launcher) launcher.setAttribute("aria-expanded", "false");
      unlockMobileScroll();
      send("host-close");
      if (launcher) launcher.focus({ preventScroll: true });
    }

    function handleDocumentClick(event) {
      if (options.mode !== "custom" || !(event.target instanceof Element)) return;
      try {
        if (!event.target.closest(options.trigger)) return;
      } catch (error) {
        return;
      }
      event.preventDefault();
      openWidget();
    }

    function handleKeyDown(event) {
      if (event.key === "Escape" && open) closeWidget();
    }

    function handleMessage(event) {
      if (!iframe || event.origin !== scriptOrigin || event.source !== iframe.contentWindow) return;
      var message = event.data;
      if (!message || message.channel !== CHANNEL || message.version !== VERSION || message.instanceId !== instanceId) return;
      if (message.event === "ready") {
        frameReady = true;
        sendIdentityCommand();
      }
      if (message.event === "identity-ready" || message.event === "identity-error" || message.event === "identity-cleared") {
        var acknowledgedRequestId = String(message.payload && message.payload.requestId || "");
        if (acknowledgedRequestId === String(identityCommandRevision)) {
          pendingIdentityAssertion = null;
          hasPendingIdentityCommand = false;
        }
      }
      if (message.event === "close") closeWidget();
      if (message.event === "escape") closeWidget();
      if (message.event === "resize" && options.mode === "inline") {
        var height = Number(message.payload && message.payload.height);
        if (Number.isFinite(height)) panel.style.height = Math.min(1200, Math.max(360, height)) + "px";
      }
      if (message.event === "open-external") {
        try {
          var href = new URL(String(message.payload && message.payload.href), scriptOrigin);
          if (href.origin === scriptOrigin) window.open(href.href, "_blank", "noopener,noreferrer");
        } catch (error) {}
      }
    }

    function destroy() {
      if (destroyed) return;
      destroyed = true;
      unlockMobileScroll();
      document.removeEventListener("click", handleDocumentClick);
      document.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("message", handleMessage);
      if (themeMedia && typeof themeMedia.removeEventListener === "function") {
        themeMedia.removeEventListener("change", syncShellTheme);
      }
      host.remove();
      if (active && active.instanceId === instanceId) active = null;
    }

    document.addEventListener("click", handleDocumentClick);
    document.addEventListener("keydown", handleKeyDown);
    window.addEventListener("message", handleMessage);
    if (options.mode === "inline") ensureFrame();

    return {
      close: closeWidget,
      destroy: destroy,
      identify: identify,
      instanceId: instanceId,
      open: openWidget
    };
  }

  var api = {
    init: function (options) {
      if (active) active.destroy();
      active = createWidget(options);
      return active;
    },
    open: function () { if (active) active.open(); },
    close: function () { if (active) active.close(); },
    destroy: function () { if (active) active.destroy(); },
    identify: function (assertion) {
      if (active) {
        active.identify(assertion);
        return;
      }
      setIdentityCommand(assertion);
    }
  };
  window.FortyOneFeedback = api;

  var automaticOptions = readOptions(currentScript);
  if (automaticOptions.portalSlug) api.init(automaticOptions);
  if (previous && Array.isArray(previous.q)) {
    previous.q.forEach(function (command) {
      var name = command && command[0];
      if (name && typeof api[name] === "function") api[name].apply(api, command.slice(1));
    });
  }
})(window, document);`;
