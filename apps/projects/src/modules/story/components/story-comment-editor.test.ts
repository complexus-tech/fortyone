/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { generateHTML } from "@tiptap/core";
import {
  getStoryCommentEditorExtensions,
  serializeStoryCommentToGitHubMarkdown,
} from "./story-comment-editor";

const MENTION = {
  id: "7d3a62e4-a7cc-4489-99b6-a986195525c5",
  label: "Thulani Museta",
};
const COMMENT_DOCUMENT = {
  content: [
    {
      content: [
        {
          attrs: {
            ...MENTION,
            mentionSuggestionChar: "@",
          },
          type: "mention",
        },
      ],
      type: "paragraph",
    },
  ],
  type: "doc",
};

jest.mock("ui", () => ({
  Avatar: "div",
  Box: "div",
  Text: "span",
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: null }),
}));

describe("story comment editor mentions", () => {
  it("renders a selected person with the mention trigger", () => {
    const html = generateHTML(
      COMMENT_DOCUMENT,
      getStoryCommentEditorExtensions({
        enableMentions: true,
        placeholder: "Leave a comment...",
      }),
    );

    const container = document.createElement("div");
    container.innerHTML = html;
    const mention = container.querySelector('a[data-type="mention"]');

    expect(mention).toHaveAttribute("data-id", MENTION.id);
    expect(mention).toHaveAttribute("data-label", MENTION.label);
    expect(mention).toHaveAttribute("href", `/profile/${MENTION.id}`);
    expect(mention).toHaveTextContent(`@${MENTION.label}`);
    expect(mention).not.toHaveTextContent("undefined");
  });

  it("serializes a selected person to GitHub-compatible Markdown", () => {
    expect(serializeStoryCommentToGitHubMarkdown(COMMENT_DOCUMENT)).toBe(
      `@${MENTION.label}`,
    );
  });
});
