import { publicDocumentHTML } from "./public-document-html";

describe("public document rendering", () => {
  it("keeps safe text highlights", () => {
    const html = publicDocumentHTML(
      '<p><mark style="background-color: #FDE68A">Important</mark></p>',
      "token",
      "https://api.example",
    );
    expect(html).toContain(
      '<mark style="background-color:#FDE68A">Important</mark>',
    );
  });

  it("keeps safe text colors and removes unrelated inline styles", () => {
    const html = publicDocumentHTML(
      '<p><span style="color: #2563EB; position: fixed">Blue note</span></p>',
      "token",
      "https://api.example",
    );

    expect(html).toContain('<span style="color:#2563EB">Blue note</span>');
    expect(html).not.toContain("position");
  });

  it("removes executable HTML and unsafe URLs", () => {
    const html = publicDocumentHTML(
      '<script>alert(1)</script><p onclick="alert(1)">Plan</p><a href="javascript:alert(1)">Link</a><img src="x" onerror="alert(1)"><iframe src="https://evil.example"></iframe>',
      "token",
      "https://api.example",
    );
    expect(html).not.toMatch(/script|onclick|onerror|javascript:|iframe/);
    expect(html).toContain("Plan");
    expect(html).toContain('rel="noopener noreferrer"');
  });
  it("routes document images through revocable public media authorization", () => {
    const html = publicDocumentHTML(
      '<img src="https://api.example/workspaces/private-team/documents/11111111-1111-1111-1111-111111111111/media/22222222-2222-2222-2222-222222222222" data-width="50%">',
      "public-token",
      "https://api.example",
    );
    expect(html).toContain(
      "https://api.example/public/documents/public-token/media/22222222-2222-2222-2222-222222222222",
    );
    expect(html).not.toContain("private-team");
  });
  it("preserves tables and renders task checkboxes as disabled", () => {
    const html = publicDocumentHTML(
      '<table><tr><td colspan="2">Plan</td></tr></table><input type="checkbox" checked><input type="password" value="secret">',
      "token",
      "https://api.example",
    );
    expect(html).toContain('<div class="tableWrapper"><table>');
    expect(html).toContain('colspan="2"');
    expect(html).toContain("disabled");
    expect(html).not.toContain("password");
    expect(html).not.toContain("secret");
  });

  it("repairs table and list markers left by legacy Markdown pastes", () => {
    const html = publicDocumentHTML(
      "<ul><li><p>Keep context close.</p></li><li><p>- Make decisions scannable.</p></li></ul><ol><li><p>First risk</p></li><li><p>2. Second risk</p></li></ol><p>| Workstream | Status |</p><p>| --- | --- |</p><p>| Editor | Ready |</p>",
      "token",
      "https://api.example",
    );

    expect(html).toContain("<p>Make decisions scannable.</p>");
    expect(html).toContain("<p>Second risk</p>");
    expect(html).toContain("<table>");
    expect(html).toContain('<div class="tableWrapper"><table>');
    expect(html).toContain("<th>Workstream</th>");
    expect(html).toContain("<td>Ready</td>");
    expect(html).not.toContain("| --- | --- |");
  });
});
