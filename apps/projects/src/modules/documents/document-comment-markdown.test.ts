import {
  hasDocumentMayaMention,
  renderDocumentComment,
} from "./document-comment-markdown";

test("renders the supported Markdown while preserving plain text and line breaks", () => {
  const html = renderDocumentComment(
    "Make **outcomes** concrete, *please*.\nNext line\n\n- One\n- Two\n\n`code`\n\n[Reference](https://example.com)",
  );
  expect(html).toContain("<strong>outcomes</strong>");
  expect(html).toContain("<em>please</em>");
  expect(html).toContain("<br />Next line");
  expect(html).toContain("<ul>");
  expect(html).toContain("<code>code</code>");
  expect(html).toContain('href="https://example.com"');
  expect(renderDocumentComment("Cost < budget & time > estimate")).toContain(
    "Cost &lt; budget &amp; time &gt; estimate",
  );
});

test("preserves typed mention identities, escapes labels, and never activates literal markup", () => {
  const id = "1d4f475b-3dfa-458f-a315-0513149e481a";
  const html = renderDocumentComment(
    `[@Joseph](mention:${id}) and [@Maya](agent:maya)`,
  );
  expect(html).toContain(`data-id="${id}"`);
  expect(html).toContain('data-agent-key="maya"');
  expect(renderDocumentComment(`[@<img src=x>](mention:${id})`)).not.toContain(
    "<img",
  );
  expect(renderDocumentComment("`[@Maya](agent:maya)`")).not.toContain(
    'data-agent-key="maya"',
  );
  expect(renderDocumentComment("[unknown](agent:other)")).not.toContain(
    'data-type="mention"',
  );
});

test("shows optimistic Maya feedback only for actionable mentions, excluding code and quotations", () => {
  expect(
    hasDocumentMayaMention("Please **ask [@Maya](agent:maya)** here."),
  ).toBe(true);
  expect(hasDocumentMayaMention("- [@Maya](agent:maya) please review")).toBe(
    true,
  );
  expect(hasDocumentMayaMention("> [@Maya](agent:maya) please review")).toBe(
    false,
  );
  expect(hasDocumentMayaMention("`[@Maya](agent:maya)`")).toBe(false);
  expect(hasDocumentMayaMention("```\n[@Maya](agent:maya)\n```")).toBe(false);
  expect(hasDocumentMayaMention("@Maya is a plain name")).toBe(false);
});

test("treats raw HTML as text and does not allow images, scripts or unsafe links", () => {
  const html = renderDocumentComment(
    '<script>alert(1)</script>\n\n<img src="https://tracker.test/image" onerror="alert(1)">\n\n![diagram](https://tracker.test/image)\n\n[bad](javascript:alert%281%29)',
  );
  expect(html).not.toMatch(/<(?:script|img|iframe)\b/);
  expect(html).not.toContain('href="javascript:');
  expect(html).toContain("&lt;script&gt;");
  expect(html).toContain("diagram");
});
