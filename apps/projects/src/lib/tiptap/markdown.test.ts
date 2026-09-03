/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { Editor } from "@tiptap/core";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import { Slice } from "@tiptap/pm/model";
import StarterKit from "@tiptap/starter-kit";
import {
  getRichTextContentType,
  looksLikeMarkdown,
  RichTextMarkdown,
  RichTextMarkdownPaste,
} from "./markdown";

describe("rich-text Markdown", () => {
  it("recognizes supported block and inline Markdown without treating prose as Markdown", () => {
    expect(looksLikeMarkdown("### Checklist\n- [ ] Upload proof")).toBe(true);
    expect(looksLikeMarkdown("Paste **bold text** here")).toBe(true);
    expect(looksLikeMarkdown("A normal description with [brackets].")).toBe(
      false,
    );
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
});
