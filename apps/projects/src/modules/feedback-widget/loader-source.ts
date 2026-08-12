export const FEEDBACK_WIDGET_LOADER_SOURCE = String.raw`(function (window, document) {
  "use strict";

  var currentScript = document.currentScript;
  var scriptOrigin = currentScript && currentScript.src
    ? new URL(currentScript.src, window.location.href).origin
    : window.location.origin;
  var previous = window.FortyOneFeedback;
  var active = null;
  var CHANNEL = "fortyone:feedback-widget";
  var VERSION = 1;

  function oneOf(value, values, fallback) {
    return values.indexOf(value) === -1 ? fallback : value;
  }

  function readOptions(script) {
    var data = script ? script.dataset : {};
    return {
      defaultTab: data.defaultTab,
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
      defaultTab: oneOf(options.defaultTab, ["feedback", "roadmap"], "feedback"),
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

  function createWidget(rawOptions) {
    var options = normalizeOptions(rawOptions || {});
    var instanceId = makeInstanceId();
    var iframe = null;
    var launcher = null;
    var open = false;
    var destroyed = false;
    var restoreOverflow = null;
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
      ".root{font-family:-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif;position:relative;z-index:2147483000}",
      ".launcher{position:fixed;bottom:24px;width:54px;height:54px;border:0;border-radius:18px;display:grid;place-items:center;background:#171717;color:#fff;box-shadow:0 16px 40px rgba(0,0,0,.24),inset 0 0 0 1px rgba(255,255,255,.12);cursor:pointer;transition:transform .2s ease,box-shadow .2s ease;z-index:2}",
      ".launcher:hover{transform:translateY(-2px);box-shadow:0 20px 44px rgba(0,0,0,.3),inset 0 0 0 1px rgba(255,255,255,.16)}",
      ".launcher:focus-visible{outline:3px solid rgba(59,130,246,.55);outline-offset:3px}",
      ".launcher svg{width:24px;height:24px;transition:transform .2s ease}",
      ".root[data-open=true] .launcher svg{transform:scale(.9) rotate(3deg)}",
      ".right .launcher{right:24px}.left .launcher{left:24px}",
      ".panel{position:fixed;bottom:92px;width:min(420px,calc(100vw - 32px));height:min(680px,calc(100dvh - 112px));overflow:hidden;border:1px solid rgba(127,127,127,.24);border-radius:22px;background:#fff;box-shadow:0 28px 80px rgba(0,0,0,.24),0 4px 18px rgba(0,0,0,.1);opacity:0;visibility:hidden;transform:translateY(12px) scale(.985);transform-origin:bottom;transition:opacity .18s ease,transform .22s cubic-bezier(.2,.8,.2,1),visibility 0s linear .22s}",
      ".right .panel{right:24px}.left .panel{left:24px}",
      ".root[data-open=true] .panel{opacity:1;visibility:visible;transform:none;transition-delay:0s}",
      ".panel iframe{display:block;width:100%;height:100%;border:0;background:#fff}",
      ".inline .panel{position:relative;inset:auto;width:100%;height:640px;min-height:360px;border-radius:18px;opacity:1;visibility:visible;transform:none;box-shadow:0 8px 30px rgba(0,0,0,.1)}",
      "@media(max-width:640px){.popup .panel{inset:0;width:100vw;height:100dvh;max-height:none;border:0;border-radius:0;transform:translateY(12px)}.popup.root[data-open=true] .panel{transform:none}.launcher{bottom:16px}.right .launcher{right:16px}.left .launcher{left:16px}}",
      "@media(prefers-reduced-motion:reduce){.launcher,.panel,.launcher svg{transition:none}}"
    ].join("");

    var root = document.createElement("div");
    root.className = "root " + (options.mode === "inline" ? "inline" : "popup") + " " + (options.position === "bottom-left" ? "left" : "right");
    root.dataset.open = options.mode === "inline" ? "true" : "false";
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
      launcher.innerHTML = '<svg aria-hidden="true" viewBox="0 0 24 24" fill="none"><path d="M7.4 18.1 4 20l1-3.8A8 8 0 1 1 7.4 18Z" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/><path d="M8.5 12h7M12 8.5v7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>';
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
    destroy: function () { if (active) active.destroy(); }
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
