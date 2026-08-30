/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getCommentEditorExtensions } from "./comment-editor";

describe("comment editor extensions", () => {
  it("provides the plain rich-text contract used outside the story feature", () => {
    const names = getCommentEditorExtensions({
      placeholder: "Leave a comment...",
    }).map(({ name }) => name);

    expect(names).toEqual(
      expect.arrayContaining([
        "link",
        "placeholder",
        "starterKit",
        "taskItem",
        "taskList",
        "underline",
      ]),
    );
    expect(names).not.toContain("mention");
  });
});
