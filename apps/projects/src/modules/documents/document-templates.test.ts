import {
  documentTemplateCategories,
  documentTemplates,
} from "./document-templates";

test("every nonblank template belongs to a visible category and its searchable text matches its content", () => {
  const categories = new Set(documentTemplateCategories.map(({ id }) => id));
  expect(new Set(documentTemplates.map(({ id }) => id)).size).toBe(
    documentTemplates.length,
  );
  for (const template of documentTemplates.filter(({ id }) => id !== "blank")) {
    expect(categories.has(template.category!)).toBe(true);
    const content = document.createElement("div");
    content.innerHTML = template.contentHtml;
    expect(content.querySelectorAll("h2").length).toBeGreaterThanOrEqual(4);
    const normalize = (text: string) => text.replace(/\s+/g, " ").trim();
    // Element boundaries in HTML may omit whitespace, so compare the text of each leaf.
    for (const element of content.querySelectorAll("h2, p, th, td, li")) {
      expect(normalize(template.contentText)).toContain(
        normalize(element.textContent),
      );
    }
  }
});
