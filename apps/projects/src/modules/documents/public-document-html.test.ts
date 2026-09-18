import { publicDocumentHTML } from "./public-document-html";

describe("public document rendering", () => {
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
    expect(html).toContain('colspan="2"');
    expect(html).toContain("disabled");
    expect(html).not.toContain("password");
    expect(html).not.toContain("secret");
  });
});
