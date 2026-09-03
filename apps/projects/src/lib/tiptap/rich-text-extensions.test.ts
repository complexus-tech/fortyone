/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

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
        "documentVideo",
        "image",
        "markdown",
        "richTextMarkdownPaste",
        "slashCommand",
        "starterKit",
        "table",
        "taskList",
      ]),
    );
  });
});
