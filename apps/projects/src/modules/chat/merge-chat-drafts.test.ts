/* global describe, expect, it -- Provided by the projects test runner. */
import type { ChatDraft } from "./chat-rich-text";
import type { ChatAttachment } from "./types";
import { mergeChatDrafts } from "./merge-chat-drafts";

const empty: ChatDraft = { text: "", content: null };
const attachment: ChatAttachment = {
  id: "file-1",
  fileName: "notes.pdf",
  mediaType: "application/pdf",
  sizeBytes: 50,
  url: "",
};

describe("new-message draft transfer", () => {
  it("retains both messages and their rich formatting with a paragraph gap", () => {
    const existing: ChatDraft = {
      text: "An unfinished message",
      content: {
        type: "doc",
        content: [
          {
            type: "paragraph",
            content: [
              {
                type: "text",
                text: "An unfinished message",
                marks: [{ type: "bold" }],
              },
            ],
          },
        ],
      },
      attachments: [attachment],
    };
    const incoming: ChatDraft = {
      text: "A new thought",
      content: {
        type: "doc",
        content: [
          {
            type: "bulletList",
            content: [
              {
                type: "listItem",
                content: [
                  {
                    type: "paragraph",
                    content: [{ type: "text", text: "A new thought" }],
                  },
                ],
              },
            ],
          },
        ],
      },
    };

    expect(mergeChatDrafts(existing, incoming)).toEqual({
      text: "An unfinished message\n\nA new thought",
      content: {
        type: "doc",
        content: [
          existing.content!.content![0],
          { type: "paragraph" },
          incoming.content!.content![0],
        ],
      },
      attachments: [attachment],
    });
  });

  it("converts multiline plain drafts to paragraphs without dropping blank lines", () => {
    const merged = mergeChatDrafts(
      { text: "First\n\nSecond", content: null },
      { text: "Third", content: null },
    );

    expect(merged.text).toBe("First\n\nSecond\n\nThird");
    expect(merged.content?.content).toEqual([
      { type: "paragraph", content: [{ type: "text", text: "First" }] },
      { type: "paragraph", content: [] },
      { type: "paragraph", content: [{ type: "text", text: "Second" }] },
      { type: "paragraph" },
      { type: "paragraph", content: [{ type: "text", text: "Third" }] },
    ]);
  });

  it("recovers content-only text and mentions so the transferred draft remains sendable", () => {
    const incoming: ChatDraft = {
      text: "",
      content: {
        type: "doc",
        content: [
          {
            type: "paragraph",
            content: [
              { type: "mention", attrs: { id: "person", label: "Sarah" } },
              { type: "text", text: " please review." },
            ],
          },
        ],
      },
    };
    expect(mergeChatDrafts(empty, incoming)).toEqual({
      text: "@Sarah please review.",
      content: incoming.content,
    });
  });

  it("does not add empty paragraphs when either draft is empty", () => {
    const draft: ChatDraft = {
      text: "Keep this",
      content: null,
      attachments: [attachment],
    };
    expect(mergeChatDrafts(draft, empty)).toEqual(draft);
    expect(mergeChatDrafts(empty, draft)).toEqual(draft);
    expect(mergeChatDrafts(empty, empty)).toEqual(empty);
  });

  it("retains attachments on an otherwise empty destination and deduplicates IDs", () => {
    const second = { ...attachment, id: "file-2" };
    const merged = mergeChatDrafts(
      { ...empty, attachments: [attachment] },
      {
        text: "New message",
        content: null,
        attachments: [{ ...attachment, fileName: "changed.pdf" }, second],
      },
    );
    expect(merged.text).toBe("New message");
    expect(merged.attachments).toEqual([attachment, second]);
  });

  it("does not mutate either source draft", () => {
    const existing: ChatDraft = {
      text: "Existing",
      content: null,
      attachments: [attachment],
    };
    const incoming: ChatDraft = { text: "Incoming", content: null };
    const before = JSON.stringify({ existing, incoming });
    Object.freeze(existing);
    Object.freeze(existing.attachments);
    Object.freeze(incoming);

    mergeChatDrafts(existing, incoming);

    expect(JSON.stringify({ existing, incoming })).toBe(before);
  });
});
