/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  buildStoryLinkPreviewHtml,
  isSlackStoryLinkPreview,
  isStoryPath,
} from "./story-link-preview";

describe("story link previews", () => {
  it.each([
    "/work/PRD-571",
    "/story/PRD-571",
    "/story/d0c8baaf-d40e-4d2f-8f37-b702da402085",
    "/story/d0c8baaf-d40e-4d2f-8f37-b702da402085/story-title",
    "/art-circles/work/PRD-571",
  ])("recognizes story route %s", (pathname) => {
    expect(isStoryPath(pathname)).toBe(true);
  });

  it("does not treat unrelated routes as stories", () => {
    expect(isStoryPath("/my-work")).toBe(false);
    expect(isStoryPath("/work")).toBe(false);
    expect(isStoryPath("/feedback/PRD-571")).toBe(false);
  });

  it("only enables the minimal preview for Slack crawlers", () => {
    expect(
      isSlackStoryLinkPreview(
        "/work/PRD-571",
        "Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)",
      ),
    ).toBe(true);
    expect(isSlackStoryLinkPreview("/work/PRD-571", "Mozilla/5.0")).toBe(false);
  });

  it("only exposes the favicon and crawler controls", () => {
    const html = buildStoryLinkPreviewHtml(
      "https://art-circles.fortyone.app/favicon.ico",
    );

    expect(html).toContain(
      'rel="icon" href="https://art-circles.fortyone.app/favicon.ico"',
    );
    expect(html).toContain('content="noindex, nofollow, noarchive"');
    expect(html).not.toMatch(
      /<title|login|create an account|property="og:|name="twitter:/i,
    );
  });
});
