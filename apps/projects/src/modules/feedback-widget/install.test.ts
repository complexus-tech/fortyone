/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
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

  it("falls back from updates until public changelog data is wired", () => {
    const snippet = buildFeedbackWidgetSnippet({
      defaultTab: "updates",
      portalSlug: "city-roads",
    });

    expect(snippet).toContain('data-default-tab="feedback"');
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
});
