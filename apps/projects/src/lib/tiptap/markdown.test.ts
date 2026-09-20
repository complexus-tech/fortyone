/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { Editor } from "@tiptap/core";
import Collaboration from "@tiptap/extension-collaboration";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import { Slice } from "@tiptap/pm/model";
import StarterKit from "@tiptap/starter-kit";
import * as Y from "yjs";
import { createRichTextExtensions } from "./rich-text-extensions";
import {
  getRichTextContentType,
  looksLikeMarkdown,
  markdownToRichTextHTML,
  richTextHTMLToMarkdown,
  RichTextMarkdown,
  RichTextMarkdownPaste,
} from "./markdown";

describe("rich-text Markdown", () => {
  it("recognizes supported block and inline Markdown without treating prose as Markdown", () => {
    expect(looksLikeMarkdown("### Checklist\n- [ ] Upload proof")).toBe(true);
    expect(looksLikeMarkdown("Paste **bold text** here")).toBe(true);
    expect(looksLikeMarkdown("Name | Status\n--- | ---\nEditor | Ready")).toBe(
      true,
    );
    expect(looksLikeMarkdown("A normal description with [brackets].")).toBe(
      false,
    );
  });

  it("converts a pasted GFM table into table nodes", () => {
    const html = markdownToRichTextHTML(
      "Workstream | Owner | Status\n--- | --- | ---\nEditor | Product | In progress",
    );

    expect(html).toContain("<table");
    expect(html).toMatch(/<th[^>]*>.*Workstream.*<\/th>/);
    expect(html).toMatch(/<td[^>]*>.*In progress.*<\/td>/);
    expect(html).not.toContain("--- | ---");
  });

  it("serializes rich-text tables for Markdown downloads", () => {
    const markdown = richTextHTMLToMarkdown(
      '<div class="tableWrapper"><table><thead><tr><th>Workstream</th><th>Status</th></tr></thead><tbody><tr><td>Editor</td><td>Ready</td></tr></tbody></table></div>',
    );

    expect(markdown).toContain("| Workstream | Status |");
    expect(markdown).toMatch(/\| Editor\s+\| Ready\s+\|/);
  });

  it("only treats legacy plain descriptions as Markdown", () => {
    expect(getRichTextContentType("### Checklist", null)).toBe("markdown");
    expect(
      getRichTextContentType("### Checklist", "<p>### Checklist</p>"),
    ).toBe("html");
    expect(getRichTextContentType("Normal description", null)).toBe("html");
  });

  it("parses GitHub-style checkboxes into interactive task items", () => {
    const editor = new Editor({
      extensions: [StarterKit, TaskList, TaskItem, RichTextMarkdown],
    });

    editor.commands.setContent(
      "### Checklist\n\n- [ ] Upload proof\n- [x] Link proof",
      { contentType: "markdown" },
    );

    const document = editor.getJSON();
    expect(document.content[0]).toMatchObject({
      attrs: { level: 3 },
      type: "heading",
    });
    expect(document.content[1]).toMatchObject({
      content: [
        { attrs: { checked: false }, type: "taskItem" },
        { attrs: { checked: true }, type: "taskItem" },
      ],
      type: "taskList",
    });

    editor.destroy();
  });

  it("converts Markdown from the plain-text clipboard on paste", () => {
    const editor = new Editor({
      extensions: [
        StarterKit,
        TaskList,
        TaskItem,
        RichTextMarkdown,
        RichTextMarkdownPaste,
      ],
    });
    const preventDefault = jest.fn();
    const event = {
      clipboardData: {
        getData: () => "- [ ] Pasted task",
      },
      preventDefault,
    } as unknown as ClipboardEvent;

    const handled = editor.view.someProp("handlePaste", (handler) =>
      handler(editor.view, event, Slice.empty),
    );

    expect(handled).toBe(true);
    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(editor.getJSON().content[0]).toMatchObject({
      content: [{ attrs: { checked: false }, type: "taskItem" }],
      type: "taskList",
    });

    editor.destroy();
  });

  it("preserves aligned tables and the content that follows them on paste", () => {
    const collaborationDocument = new Y.Doc();
    const richTextExtensions = createRichTextExtensions({
      collaborative: true,
      onMediaFiles: () => {},
      onMediaRequest: () => {},
      placeholder: "Type / for commands",
    });
    const editor = new Editor({
      extensions: [
        ...richTextExtensions,
        Collaboration.configure({ document: collaborationDocument }),
      ],
    });
    const preventDefault = jest.fn();
    const markdown = `## Workstream status

| Workstream | Owner | Status |
| :--------- | :---- | :----- |
| Editor | Design | In progress |

## Alignment test

| Item | Quantity | Progress |
| :--- | -------: | :------: |
| Editor polish | 4 | 75% |

## Final checks

Content after tables must remain intact.`;
    const event = {
      clipboardData: { getData: () => markdown },
      preventDefault,
    } as unknown as ClipboardEvent;

    const handled = editor.view.someProp("handlePaste", (handler) =>
      handler(editor.view, event, Slice.empty),
    );
    const document = editor.getJSON();

    expect(handled).toBe(true);
    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(
      document.content.filter((node) => node.type === "table"),
    ).toHaveLength(2);
    expect(editor.getText()).toContain("Workstream");
    expect(editor.getText()).toContain("Editor polish");
    expect(editor.getText()).toContain("Final checks");
    expect(editor.getText()).toContain(
      "Content after tables must remain intact.",
    );

    editor.destroy();
    collaborationDocument.destroy();
  });
});
