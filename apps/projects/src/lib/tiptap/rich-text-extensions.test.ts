/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { createRichTextExtensions } from "./rich-text-extensions";

describe("rich-text editor extensions", () => {
  it("registers every node used by the shared slash-command experience", () => {
    const names = createRichTextExtensions({
      onMediaFiles: () => undefined,
      onMediaRequest: () => undefined,
      placeholder: "Type / for commands",
    }).map(({ name }) => name);

    expect(names).toEqual(
      expect.arrayContaining([
        "blockquote",
        "bulletList",
        "codeBlock",
        "documentVideo",
        "heading",
        "horizontalRule",
        "image",
        "orderedList",
        "slashCommand",
        "table",
        "taskList",
      ]),
    );
  });
});
