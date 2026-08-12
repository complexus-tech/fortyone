/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { FEEDBACK_WIDGET_LOADER_SOURCE } from "./loader-source";

describe("feedback widget loader", () => {
  it("exposes the lifecycle API and lazy iframe path", () => {
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("init: function");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("open: function");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("close: function");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("destroy: function");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      '"/embed/feedback/" + encodeURIComponent(options.portalSlug)',
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain(
      'if (options.mode === "inline") ensureFrame()',
    );
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain("(?=.{3,255}$)");
    expect(FEEDBACK_WIDGET_LOADER_SOURCE).toContain('["feedback", "roadmap"]');
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
      'if (message.event === "escape") closeWidget()',
    );
  });
});
