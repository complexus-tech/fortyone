/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { Editor } from "@tiptap/core";
import { createDocumentExtensions } from "@fortyone/document-editor";
import { createRichTextExtensions } from "./rich-text-extensions";

describe("rich-text editor extensions", () => {
  it("registers the starter kit and every separately configured extension", () => {
    const names = createRichTextExtensions({
      onMediaFiles: () => undefined,
      onMediaRequest: () => undefined,
      placeholder: "Type / for commands",
    }).map(({ name }) => name);

    expect(names).toEqual(
      expect.arrayContaining([
        "color",
        "documentVideo",
        "highlight",
        "image",
        "markdown",
        "richTextMarkdownPaste",
        "slashCommand",
        "starterKit",
        "table",
        "taskList",
        "textStyle",
      ]),
    );
  });

  it("applies text color and highlight to a selection", () => {
    const editor = new Editor({
      content: "Color this text",
      extensions: createDocumentExtensions(),
    });

    editor
      .chain()
      .setTextSelection({ from: 1, to: 6 })
      .setColor("#2563EB")
      .setHighlight({ color: "#FDE68A" })
      .run();

    expect(editor.getHTML()).toContain("color: #2563EB");
    expect(editor.getHTML()).toContain("background-color: #FDE68A");
    editor.destroy();
  });
});
