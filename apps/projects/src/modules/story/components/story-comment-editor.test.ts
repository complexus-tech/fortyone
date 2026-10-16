/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { Editor, generateHTML } from "@tiptap/core";
import {
  extractAgentMentionsFromHTML,
  extractMentionsFromHTML,
} from "@/lib/utils/mentions";
import type { MentionItem } from "./mentions/list";
import {
  getStoryCommentEditorExtensions,
  serializeStoryCommentToGitHubMarkdown,
  setStoryCommentMentionUsers,
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
type MentionItems = (input: { editor: Editor; query: string }) => MentionItem[];
type MentionExtensionOptions = {
  suggestion: { items: MentionItems };
};

jest.mock("ui", () => ({
  Avatar: "div",
  Box: "div",
  Text: "span",
}));
jest.mock("./comment-attachment-list", () => ({
  CommentAttachmentList: () => null,
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: null }),
}));

describe("story comment editor mentions", () => {
  it("round-trips a typed Maya mention alongside a human mention without treating Maya as a person", () => {
    const extensions = getStoryCommentEditorExtensions({
      enableMentions: true,
      placeholder: "Comment",
    });
    const editor = new Editor({
      extensions,
      content: {
        type: "doc",
        content: [
          {
            type: "paragraph",
            content: [
              {
                type: "mention",
                attrs: {
                  ...MENTION,
                  mentionKind: "agent",
                  agentKey: "maya",
                  label: "Maya",
                },
              },
              { type: "text", text: " ask " },
              { type: "mention", attrs: { id: "human", label: "Maya Smith" } },
            ],
          },
        ],
      },
    });
    const html = editor.getHTML();
    expect(html).toContain('data-mention-kind="agent"');
    expect(html).toContain('data-agent-key="maya"');
    const restored = new Editor({ extensions, content: html });
    expect(restored.getHTML()).toBe(html);
    expect(extractAgentMentionsFromHTML(html)).toEqual(["maya"]);
    expect(extractMentionsFromHTML(html)).toEqual([
      { id: "human", fullName: "Maya Smith" },
    ]);
    const container = document.createElement("div");
    container.innerHTML = html;
    expect(
      container.querySelector('[data-mention-kind="agent"]'),
    ).not.toHaveAttribute("href");
    editor.destroy();
    restored.destroy();
  });

  it("does not invoke Maya from plain text, quoted mentions, or code", () => {
    const mention = `<span data-type="mention" data-mention-kind="agent" data-agent-key="maya" data-id="${MENTION.id}">@Maya</span>`;
    expect(
      extractAgentMentionsFromHTML("<p>@Maya please close this</p>"),
    ).toEqual([]);
    expect(
      extractAgentMentionsFromHTML(
        `<blockquote>${mention}</blockquote><pre>${mention}</pre>`,
      ),
    ).toEqual([]);
    expect(extractAgentMentionsFromHTML(mention + mention)).toEqual(["maya"]);
  });

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

  it("uses team members that load after the editor is initialized", () => {
    const editor = new Editor({
      extensions: getStoryCommentEditorExtensions({
        enableMentions: true,
        placeholder: "Leave a comment...",
      }),
    });
    const mentionExtension = editor.extensionManager.extensions.find(
      (extension) => extension.name === "mention",
    );
    const mentionOptions = mentionExtension?.options as
      | MentionExtensionOptions
      | undefined;
    if (!mentionOptions) {
      throw new Error("Mention extension is not configured");
    }
    const getItems = mentionOptions.suggestion.items;

    expect(getItems({ editor, query: "" })).toEqual([]);

    const mentionUsers: MentionItem[] = [
      {
        id: MENTION.id,
        label: MENTION.label,
        username: "thulani",
      },
    ];
    setStoryCommentMentionUsers(editor, mentionUsers);

    expect(getItems({ editor, query: "thul" })).toEqual(mentionUsers);

    editor.destroy();
  });
});
