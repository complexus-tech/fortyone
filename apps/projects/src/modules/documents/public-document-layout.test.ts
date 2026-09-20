import { createStandaloneDocumentHTML } from "./public-document-export";

describe("public document downloads", () => {
  it("creates a complete standalone HTML document", () => {
    const html = createStandaloneDocumentHTML({
      contentHtml: "<p>Release notes</p><table><tr><td>Ready</td></tr></table>",
      title: "Q4 <Plan>",
      updatedAt: "September 20, 2026",
    });

    expect(html).toContain("<!doctype html>");
    expect(html).toContain("<title>Q4 &lt;Plan&gt;</title>");
    expect(html).toContain("<p>Release notes</p>");
    expect(html).toContain("<table>");
    expect(html).toContain(
      ".content img, .content video { border: 1px solid #e5e5e5; border-radius: 13px; corner-shape: squircle;",
    );
    expect(html).toContain("Updated September 20, 2026");
  });
});
