import assert from "node:assert/strict";
import { readFileSync, readdirSync } from "node:fs";
import test from "node:test";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { compileMDX } from "next-mdx-remote/rsc";
import matter from "gray-matter";
import { createArticleMarkdown } from "../app/(marketing)/blog/[slug]/article-markdown.ts";

void test("rich Markdown renders tables, task lists, linked duplicate headings, and themed code", async () => {
  const markdown = createArticleMarkdown();
  const { content } = await compileMDX({
    source:
      "## A heading\n\n## A heading\n\n| View | Purpose |\n| --- | --- |\n| Board | Flow |\n\n- [x] Reviewed\n\n```js\nconst value = 42;\n```",
    options: markdown.options,
  });
  const html = renderToStaticMarkup(content);
  assert.deepEqual(
    markdown.headings.map(({ id }) => id),
    ["a-heading", "a-heading-1"],
  );
  for (const heading of markdown.headings)
    assert.ok(html.includes(`href="#${heading.id}"`));
  assert.match(html, /<table>/);
  assert.match(html, /type="checkbox"/);
  assert.match(html, /--shiki-light/);
  assert.match(html, /--shiki-dark/);
});

void test("every published guide compiles its visual components without JavaScript expressions", async () => {
  const directory = new URL("../content/blog/", import.meta.url);
  const filenames = readdirSync(directory).filter((name) =>
    name.endsWith(".mdx"),
  );
  await Promise.all(
    filenames.map(async (filename) => {
      const { content: source } = matter(
        readFileSync(new URL(filename, directory), "utf8"),
      );
      const markdown = createArticleMarkdown();
      const seen = new Map<string, number>();
      const components = Object.fromEntries(
        [
          "Takeaway",
          "ProductFigure",
          "WorkflowSteps",
          "WorkflowStep",
          "ArticleChecklist",
          "ChecklistItem",
          "ArticleDetails",
        ].map((name) => [
          name,
          (props: {
            children?: React.ReactNode;
            title?: string;
            light?: string;
          }) => {
            seen.set(name, (seen.get(name) ?? 0) + 1);
            if (name === "ProductFigure")
              assert.ok(props.light?.startsWith("/images/product/"), filename);
            if (["WorkflowStep", "ChecklistItem"].includes(name))
              assert.ok(props.children, `${filename}: empty ${name}`);
            return createElement("div", null, props.children);
          },
        ]),
      );
      const { content } = await compileMDX({
        source,
        components,
        options: markdown.options,
      });
      const html = renderToStaticMarkup(content);
      assert.equal(seen.get("WorkflowStep"), 3, filename);
      assert.equal(seen.get("ChecklistItem"), 4, filename);
      assert.equal(seen.get("ProductFigure"), 1, filename);
      assert.ok(markdown.headings.length >= 6, filename);
      for (const heading of markdown.headings)
        assert.ok(html.includes(`id="${heading.id}"`), filename);
    }),
  );
});
