/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { FEEDBACK_WIDGET_LOADER_SOURCE } from "./loader-source";

type FeedbackWidgetApi = {
  destroy: () => void;
  identify: (assertion?: string | null) => void;
  init: (options: {
    portalSlug: string;
    theme?: "auto" | "dark" | "light";
  }) => {
    identify: (assertion?: string | null) => void;
    open: () => void;
  };
  open: () => void;
};

const originalCurrentScript = Object.getOwnPropertyDescriptor(
  document,
  "currentScript",
);
const originalMatchMedia = Object.getOwnPropertyDescriptor(
  window,
  "matchMedia",
);

const loadWidget = (
  options: { prefersDark?: boolean; theme?: "auto" | "dark" | "light" } = {},
) => {
  const script = document.createElement("script");
  script.dataset.portal = "city-roads";
  if (options.theme) script.dataset.theme = options.theme;
  script.src = "https://feedback.example.com/widget.js";
  document.body.appendChild(script);
  Object.defineProperty(document, "currentScript", {
    configurable: true,
    value: script,
  });
  // eslint-disable-next-line no-eval -- Execute the production loader source in its real browser-global shape.
  window.eval(FEEDBACK_WIDGET_LOADER_SOURCE);
  const api = (window as unknown as { FortyOneFeedback: FeedbackWidgetApi })
    .FortyOneFeedback;
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: jest.fn(() => ({
      addEventListener: jest.fn(),
      matches: options.prefersDark ?? false,
      removeEventListener: jest.fn(),
    })),
  });
  const instance = api.init({
    portalSlug: "city-roads",
    ...(options.theme ? { theme: options.theme } : {}),
  });
  instance.open();
  const hosts = document.querySelectorAll<HTMLElement>(
    "[data-fortyone-feedback-root]",
  );
  if (hosts.length === 0) throw new Error("Widget host was not created");
  const host = hosts.item(hosts.length - 1);
  const iframe = host.shadowRoot?.querySelector("iframe");
  if (!iframe) throw new Error("Widget iframe was not created");
  const postMessage = jest.fn();
  const frameWindow = { postMessage } as unknown as Window;
  Object.defineProperty(iframe, "contentWindow", {
    configurable: true,
    value: frameWindow,
  });
  const sendFrameEvent = (
    event: "identity-cleared" | "ready",
    payload?: Record<string, unknown>,
  ) => {
    window.dispatchEvent(
      new MessageEvent("message", {
        data: {
          channel: "fortyone:feedback-widget",
          event,
          instanceId: host.shadowRoot
            ?.querySelector<HTMLElement>(".panel")
            ?.id.replace("fortyone-feedback-panel-", ""),
          payload,
          version: 1,
        },
        origin: "https://feedback.example.com",
        source: frameWindow,
      }),
    );
  };

  return {
    acknowledgeClear: (requestId: string) => {
      sendFrameEvent("identity-cleared", { requestId });
    },
    api,
    host,
    postMessage,
    ready: () => {
      sendFrameEvent("ready");
    },
  };
};

afterEach(() => {
  (
    window as unknown as { FortyOneFeedback?: FeedbackWidgetApi }
  ).FortyOneFeedback?.destroy();
  document.body.replaceChildren();
  delete (window as unknown as { FortyOneFeedback?: FeedbackWidgetApi })
    .FortyOneFeedback;
  if (originalCurrentScript) {
    Object.defineProperty(document, "currentScript", originalCurrentScript);
  } else {
    delete (document as unknown as { currentScript?: HTMLScriptElement })
      .currentScript;
  }
  if (originalMatchMedia) {
    Object.defineProperty(window, "matchMedia", originalMatchMedia);
  } else {
    delete (window as unknown as { matchMedia?: Window["matchMedia"] })
      .matchMedia;
  }
  jest.restoreAllMocks();
});

describe("feedback widget loader", () => {
  it("exposes the lifecycle API and lazy iframe path", () => {
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("init: function");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("open: function");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("close: function");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("destroy: function");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("identify: function");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      '"/embed/feedback/" + encodeURIComponent(options.portalSlug)',
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      'if (options.mode === "inline") ensureFrame()',
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("(?=.{3,255}$)");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      '["home", "feedback", "roadmap", "updates"]',
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain('], "home"),');
  });

  it("validates message origin, source, version, and instance", () => {
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      "event.origin !== scriptOrigin || event.source !== iframe.contentWindow",
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      "message.version !== VERSION || message.instanceId !== instanceId",
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      "iframe.contentWindow.postMessage",
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).not.toContain(
      'postMessage(message, "*"',
    );
  });

  it("does not expose anonymous receipt capabilities to the host page", () => {
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).not.toContain(
      'if (message.event === "receipt")',
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).not.toContain(
      "feedback-widget:receipts",
    );
  });

  it("keeps signed identity in memory and sends it only after frame readiness", () => {
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      'if (message.event === "ready")',
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      'send("host-identify", { assertion: pendingIdentityAssertion, requestId: requestId })',
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      "var pendingIdentityAssertion = null",
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).not.toContain("localStorage");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).not.toContain("sessionStorage");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).not.toContain(
      'params.set("assertion"',
    );
  });

  it("supports floating, custom-trigger, inline, and mobile-sheet modes", () => {
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      '["bubble", "custom", "inline"]',
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      "event.target.closest(options.trigger)",
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("@media(max-width:640px)");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("height:100dvh");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      "height:min(680px,calc(100dvh - 98px))",
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("border-radius:.825rem");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("--widget-border:#292824");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      "background:var(--widget-background)",
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      "@supports(corner-shape:squircle){.panel{border-radius:2.2rem;corner-shape:squircle}}",
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      'if (message.event === "escape") closeWidget()',
    );
  });

  it("applies a dark shell before the iframe reports that it is ready", () => {
    const { host } = loadWidget({ theme: "dark" });

    expect(
      host.shadowRoot?.querySelector<HTMLElement>(".root")?.dataset.theme,
    ).toBe("dark");
    expect(host.shadowRoot?.querySelector("style")?.textContent).toContain(
      ".root[data-theme=dark]",
    );
  });

  it("delivers an explicit identity clear on logout, including before readiness", () => {
    const { api, postMessage, ready } = loadWidget();

    api.identify("identity-a");
    api.identify(null);
    ready();

    expect(postMessage).toHaveBeenCalledTimes(1);
    expect(postMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        event: "host-identity-clear",
        payload: { requestId: "2" },
      }),
      "https://feedback.example.com",
    );
  });

  it("sends each A-to-B identity command once and never replays A", () => {
    const { api, postMessage, ready } = loadWidget();
    ready();

    api.identify("identity-a");
    api.identify("identity-b");
    ready();

    const identityCalls = postMessage.mock.calls.filter(
      ([message]) => (message as { event?: string }).event === "host-identify",
    );
    expect(identityCalls).toHaveLength(2);
    expect(identityCalls[0]?.[0]).toEqual(
      expect.objectContaining({
        payload: { assertion: "identity-a", requestId: "1" },
      }),
    );
    expect(identityCalls[1]?.[0]).toEqual(
      expect.objectContaining({
        payload: { assertion: "identity-b", requestId: "2" },
      }),
    );
  });

  it("retries an unacknowledged clear and stops after the matching ack", () => {
    const { acknowledgeClear, api, postMessage, ready } = loadWidget();
    ready();
    api.identify(null);
    ready();
    acknowledgeClear("1");
    ready();

    const clearCalls = postMessage.mock.calls.filter(
      ([message]) =>
        (message as { event?: string }).event === "host-identity-clear",
    );
    expect(clearCalls).toHaveLength(2);
  });
});
