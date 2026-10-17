import { publicDocumentHTML } from "./public-document-html";

describe("public document rendering", () => {
  it("renders saved work references as titles without exposing navigation metadata", () => {
    const html = publicDocumentHTML(
      '<p>Ship <span data-type="document-reference" data-entity-type="story" data-id="11111111-1111-1111-1111-111111111111" data-display-id="PRO-98" data-label="Refine onboarding">PRO-98 Refine onboarding</span> toward <span data-type="document-reference" data-entity-type="objective" data-display-id="OBJ-7" data-label="Improve activation">OBJ-7 Improve activation</span>.</p>',
      "token",
      "https://api.example",
    );

    expect(html).toBe(
      "<p>Ship <span>Refine onboarding</span> toward <span>Improve activation</span>.</p>",
    );
  });

  it("escapes reference titles and preserves text when a legacy reference has no title attribute", () => {
    const html = publicDocumentHTML(
      '<p><span data-type="document-reference" data-label="&lt;img src=x onerror=alert(1)&gt; &amp; title">PRO-98 Old title</span><span data-type="document-reference">Existing reference</span></p>',
      "token",
      "https://api.example",
    );

    expect(html).toBe(
      "<p><span>&lt;img src=x onerror=alert(1)&gt; &amp; title</span><span>Existing reference</span></p>",
    );
    expect(html).not.toContain("<img");
  });

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

  it("preserves task labels, checked state, and nested content without enabling public controls", () => {
    const html = publicDocumentHTML(
      '<ul data-type="taskList"><li data-type="taskItem" data-checked="true"><label contenteditable="false" onclick="alert(1)"><input type="checkbox" checked><span></span></label><div><p>Headings render as headings</p><ul data-type="taskList"><li data-type="taskItem" data-checked="false"><label><input type="checkbox"><span></span></label><div><p>Nested task</p><ul><li><p>Supporting detail</p></li></ul></div></li></ul></div></li></ul>',
      "token",
      "https://api.example",
    );
    const content = document.createElement("div");
    content.innerHTML = html;
    const tasks = content.querySelectorAll('li[data-type="taskItem"]');

    expect(tasks).toHaveLength(2);
    tasks.forEach((task, index) => {
      const checkbox = task.querySelector<HTMLInputElement>(
        ':scope > label > input[type="checkbox"]',
      );

      expect(checkbox).not.toBeNull();
      expect(checkbox).toBeDisabled();
      expect(checkbox?.checked).toBe(index === 0);
      expect(task.querySelector(":scope > div > p")).not.toBeNull();
    });
    expect(
      tasks[0].querySelector(':scope > div > ul[data-type="taskList"]'),
    ).toContainElement(tasks[1] as HTMLElement);
    expect(
      tasks[1].querySelector(":scope > div > ul > li > p"),
    ).toHaveTextContent("Supporting detail");
    expect(html).not.toMatch(/onclick|contenteditable/);
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
