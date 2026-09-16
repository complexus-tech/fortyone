import assert from "node:assert/strict";
import test from "node:test";
import { serializeCommentToGitHubMarkdown } from "./comment-markdown";

test("provider and objective comments preserve mentions, formatting and task lists", () => {
  const markdown = serializeCommentToGitHubMarkdown({
    type: "doc",
    content: [
      {
        type: "paragraph",
        content: [
          { type: "mention", attrs: { label: "Rudo" } },
          { type: "text", text: " ready", marks: [{ type: "bold" }] },
        ],
      },
      {
        type: "taskList",
        content: [
          {
            type: "taskItem",
            attrs: { checked: true },
            content: [
              {
                type: "paragraph",
                content: [{ type: "text", text: "Ship the update" }],
              },
            ],
          },
        ],
      },
    ],
  });
  assert.equal(markdown, "@Rudo** ready**\n\n- [x] Ship the update");
});

test("comment serialization escapes literal syntax and preserves paragraph breaks", () => {
  const markdown = serializeCommentToGitHubMarkdown({
    type: "doc",
    content: [
      {
        type: "paragraph",
        content: [
          { type: "text", text: "[literal]" },
          { type: "hardBreak" },
          { type: "text", text: "Next line" },
        ],
      },
      { type: "paragraph", content: [{ type: "text", text: "Follow-up" }] },
    ],
  });
  assert.equal(markdown, "\\[literal\\]\\\nNext line\n\nFollow-up");
});
