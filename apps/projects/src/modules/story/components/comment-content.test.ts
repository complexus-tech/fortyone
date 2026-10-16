/* global describe, expect, it -- Jest globals are provided by the projects test runner. */
import { renderCommentHTML } from "./comment-content";

describe("comment content", () => {
  it("renders legacy Markdown wrapped in otherwise plain paragraphs", () => {
    expect(
      renderCommentHTML(
        "<p>Suggested shorter title: **Refine Onboarding Conversation**</p>",
      ),
    ).toContain("<strong>Refine Onboarding Conversation</strong>");
    const html = renderCommentHTML(
      "## Next steps\n\n- **Review** copy\n- Close the ticket",
    );
    expect(html).toContain("<h2>Next steps</h2>");
    expect(html).toContain("<ul>");
    expect(html).toContain("<strong>Review</strong>");
  });
  it("preserves rich HTML, literal code, and structured mentions without double conversion", () => {
    const html =
      '<p><strong>Already bold</strong> **literal** <span class="mention" data-type="mention" data-mention-kind="agent" data-agent-key="maya" data-id="maya">@Maya</span></p><pre><code>**code**</code></pre>';
    expect(renderCommentHTML(html)).toBe(html);
  });
  it("preserves safe legacy image previews while removing image event handlers", () => {
    const html = renderCommentHTML(
      '<p>Screenshot</p><img src="https://files.example.test/screenshot.png" alt="Screenshot" onerror="alert(1)">',
    );
    expect(html).toContain(
      '<img src="https://files.example.test/screenshot.png" alt="Screenshot"',
    );
    expect(html).not.toContain("onerror");
  });
  it("strips executable HTML and unsafe URLs in legacy and rich content", () => {
    const html = renderCommentHTML(
      '<p onclick="alert(1)">**Safe**</p><script>alert(1)</script><img src=x onerror="alert(1)">',
    );
    expect(html).toContain("Safe");
    expect(html).not.toMatch(/script|onclick|onerror|alert/);
    expect(
      renderCommentHTML('<p><a href="javascript:alert(1)">bad link</a></p>'),
    ).not.toMatch(/javascript:/);
    expect(
      renderCommentHTML("[unsafe](javascript:alert%281%29)"),
    ).not.toContain("href=");
  });
});
