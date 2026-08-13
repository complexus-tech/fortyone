/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  buildFeedbackWidgetIdentityServerExample,
  buildFeedbackWidgetSnippet,
  buildInlineFeedbackWidgetMarkup,
} from "./install";

describe("feedback widget install snippet", () => {
  it("builds a declarative, asynchronous bubble installation", () => {
    const snippet = buildFeedbackWidgetSnippet({
      defaultTab: "roadmap",
      portalSlug: "city-roads",
      position: "bottom-left",
      scriptOrigin: "https://cloud.fortyone.app/ignored",
      theme: "dark",
    });

    expect(snippet).toContain(
      'src="https://cloud.fortyone.app/api/feedback-widget/v1.js"',
    );
    expect(snippet).toContain('data-portal="city-roads"');
    expect(snippet).toContain('data-default-tab="roadmap"');
    expect(snippet).toContain('data-position="bottom-left"');
    expect(snippet).toContain("async");
  });

  it("adds a target for inline installations", () => {
    const snippet = buildInlineFeedbackWidgetMarkup({
      portalSlug: "city-roads",
    });

    expect(snippet).toContain('<div id="fortyone-feedback"></div>');
    expect(snippet).toContain('data-mode="inline"');
    expect(snippet).toContain('data-target="#fortyone-feedback"');
  });

  it("can open directly on published updates", () => {
    const snippet = buildFeedbackWidgetSnippet({
      defaultTab: "updates",
      portalSlug: "city-roads",
    });

    expect(snippet).toContain('data-default-tab="updates"');
  });

  it("includes the non-secret widget key identifier", () => {
    const snippet = buildFeedbackWidgetSnippet({
      portalSlug: "city-roads",
      widgetKeyId: "widget-key-123",
    });

    expect(snippet).toContain('data-key-id="widget-key-123"');
    expect(snippet).not.toContain("signing-secret");
  });

  it("escapes custom trigger selectors in HTML attributes", () => {
    const snippet = buildFeedbackWidgetSnippet({
      mode: "custom",
      portalSlug: "city-roads",
      trigger: '[data-feedback="open"]',
    });

    expect(snippet).toContain(
      'data-trigger="[data-feedback=&quot;open&quot;]"',
    );
  });

  it("builds a server-only signed identity example without a real secret", () => {
    const example = buildFeedbackWidgetIdentityServerExample({
      origin: "https://app.example.com",
      signingSecretVersion: 2,
      widgetKeyId: "widget-key-123",
    });

    expect(example).toContain('keyId: "widget-key-123"');
    expect(example).toContain("version: 2");
    expect(example).toContain("FORTYONE_FEEDBACK_WIDGET_SECRET");
    expect(example).toContain('origin: "https://app.example.com"');
    expect(example).not.toContain("data-key-id");
  });
});
