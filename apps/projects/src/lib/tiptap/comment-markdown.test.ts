/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { serializeCommentToGitHubMarkdown } from "./comment-markdown";

describe("comment Markdown serialization", () => {
  it("serializes mentions and task lists outside the Story feature", () => {
    expect(
      serializeCommentToGitHubMarkdown({
        content: [
          {
            content: [
              {
                attrs: { label: "Rudo" },
                type: "mention",
              },
            ],
            type: "paragraph",
          },
          {
            content: [
              {
                attrs: { checked: true },
                content: [
                  {
                    content: [{ text: "Ship the update", type: "text" }],
                    type: "paragraph",
                  },
                ],
                type: "taskItem",
              },
            ],
            type: "taskList",
          },
        ],
        type: "doc",
      }),
    ).toBe("@Rudo\n\n- [x] Ship the update");
  });

  it("preserves Markdown block spacing and inline hard breaks", () => {
    expect(
      serializeCommentToGitHubMarkdown({
        content: [
          {
            content: [
              { text: "First line", type: "text" },
              { type: "hardBreak" },
              { text: "Second line", type: "text" },
            ],
            type: "paragraph",
          },
          {
            content: [{ text: "Follow-up", type: "text" }],
            type: "paragraph",
          },
        ],
        type: "doc",
      }),
    ).toBe("First line\\\nSecond line\n\nFollow-up");
  });
});
