/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  buildMinimalLinkPreviewHtml,
  isMinimalLinkPreviewPath,
  isSlackMinimalLinkPreview,
} from "./story-link-preview";

describe("minimal Slack link previews", () => {
  it.each([
    "/work/PRD-571",
    "/story/PRD-571",
    "/story/d0c8baaf-d40e-4d2f-8f37-b702da402085",
    "/story/d0c8baaf-d40e-4d2f-8f37-b702da402085/story-title",
    "/art-circles/work/PRD-571",
    "/teams/9d4cb335-a293-4f42-b930-30885669c648/requests/8f09dcbb-1f04-4370-a27f-081631211322",
    "/teams/9d4cb335-a293-4f42-b930-30885669c648/requests/8f09dcbb-1f04-4370-a27f-081631211322/",
    "/art-circles/teams/9d4cb335-a293-4f42-b930-30885669c648/requests/8f09dcbb-1f04-4370-a27f-081631211322",
  ])("recognizes supported route %s", (pathname) => {
    expect(isMinimalLinkPreviewPath(pathname)).toBe(true);
  });

  it.each([
    "/my-work",
    "/work",
    "/feedback/PRD-571",
    "/teams/9d4cb335-a293-4f42-b930-30885669c648/requests",
    "/teams/not-a-team/requests/not-a-request",
    "/teams/9d4cb335-a293-4f42-b930-30885669c648/requests/8f09dcbb-1f04-4370-a27f-081631211322/activity",
  ])("does not treat unrelated route %s as a minimal preview", (pathname) => {
    expect(isMinimalLinkPreviewPath(pathname)).toBe(false);
  });

  it("only enables the minimal preview for Slack crawlers", () => {
    expect(
      isSlackMinimalLinkPreview(
        "/teams/9d4cb335-a293-4f42-b930-30885669c648/requests/8f09dcbb-1f04-4370-a27f-081631211322",
        "Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)",
      ),
    ).toBe(true);
    expect(isSlackMinimalLinkPreview("/work/PRD-571", "Mozilla/5.0")).toBe(
      false,
    );
  });

  it("only exposes the favicon and crawler controls", () => {
    const html = buildMinimalLinkPreviewHtml(
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
