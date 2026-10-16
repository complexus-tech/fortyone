/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { Editor } from "@tiptap/core";
import { Slice } from "@tiptap/pm/model";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import {
  CommentAttachment,
  CommentAttachmentDrop,
  setCommentAttachmentHandler,
} from "./comment-attachment-extension";
import {
  COMMENT_ATTACHMENT_NODE,
  commentAttachmentUrl,
  insertCommentAttachment,
  splitCommentAttachments,
} from "./comment-attachments";

jest.mock("./comment-attachment-list", () => ({
  CommentAttachmentList: () => null,
}));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.test",
}));

const ATTACHMENT_ID = "7d3a62e4-a7cc-4489-99b6-a986195525c5";
const FILE = {
  id: ATTACHMENT_ID,
  filename: "Release <plan>.pdf",
  url: commentAttachmentUrl("acme", "story", ATTACHMENT_ID),
  mimeType: "application/pdf",
  size: 128,
  uploadedBy: "author",
  createdAt: "2026-10-16T12:00:00Z",
};
const extensions = [Document, Paragraph, Text, CommentAttachment];

describe("comment attachment persistence", () => {
  it("restores the empty editor after repeated upload and removal cycles", () => {
    const editor = new Editor({ extensions });
    for (let index = 0; index < 3; index++) {
      expect(insertCommentAttachment(editor, FILE)).toBe(true);
      expect(editor.isEmpty).toBe(false);
      const attachment = editor.state.doc.firstChild!;
      expect(attachment.type.name).toBe(COMMENT_ATTACHMENT_NODE);
      editor.commands.deleteRange({ from: 0, to: attachment.nodeSize });
      expect(editor.getHTML()).toBe("<p></p>");
      expect(editor.isEmpty).toBe(true);
    }
    editor.destroy();
  });

  it("uses the current upload handler for paste and drop and respects a locked editor", () => {
    const original = jest.fn();
    const current = jest.fn();
    const editor = new Editor({
      extensions: [
        ...extensions,
        CommentAttachmentDrop.configure({ onFiles: original }),
      ],
    });
    const file = new File(["image"], "screenshot.png", { type: "image/png" });
    const paste = new Event("paste", { cancelable: true }) as ClipboardEvent;
    Object.defineProperty(paste, "clipboardData", { value: { files: [file] } });
    setCommentAttachmentHandler(editor, current);
    editor.view.someProp("handlePaste", (handle) =>
      handle(editor.view, paste, Slice.empty),
    );
    expect(paste.defaultPrevented).toBe(true);
    expect(original).not.toHaveBeenCalled();
    expect(current).toHaveBeenCalledWith(editor, [file]);
    current.mockClear();
    editor.setEditable(false);
    const drop = new Event("drop", { cancelable: true }) as DragEvent;
    Object.defineProperty(drop, "dataTransfer", { value: { files: [file] } });
    editor.view.someProp("handleDrop", (handle) =>
      handle(editor.view, drop, Slice.empty, false),
    );
    expect(drop.defaultPrevented).toBe(false);
    expect(current).not.toHaveBeenCalled();
    editor.destroy();
  });

  it.each(["application/pdf", "image/png"])(
    "retains %s metadata through saved HTML, editing, and draft restoration",
    (mimeType) => {
      const file = { ...FILE, mimeType };
      const editor = new Editor({
        extensions,
        content: {
          type: "doc",
          content: [
            { type: "paragraph", content: [{ type: "text", text: "Review" }] },
            { type: COMMENT_ATTACHMENT_NODE, attrs: file },
          ],
        },
      });
      const html = editor.getHTML();
      const restored = new Editor({ extensions, content: html });
      expect(restored.getJSON().content[1]).toEqual({
        type: COMMENT_ATTACHMENT_NODE,
        attrs: file,
      });
      expect(splitCommentAttachments(html)).toEqual({
        html: "<p>Review</p>",
        attachments: [file],
      });
      expect(html).not.toContain("blob:");
      expect(html).not.toContain("?token=");
      expect(html.includes("<img")).toBe(mimeType.startsWith("image/"));
      editor.destroy();
      restored.destroy();
    },
  );

  it("keeps an attachment-only comment nonempty and deduplicates previews", () => {
    const editor = new Editor({
      extensions,
      content: {
        type: "doc",
        content: [{ type: COMMENT_ATTACHMENT_NODE, attrs: FILE }],
      },
    });
    expect(editor.isEmpty).toBe(false);
    const html = editor.getHTML();
    expect(splitCommentAttachments(html + html).attachments).toEqual([FILE]);
    editor.destroy();
  });

  it.each([
    `https://other.example.test/workspaces/acme/stories/story/attachments/${ATTACHMENT_ID}`,
    `${FILE.url}?token=temporary`,
    `https://api.example.test/other/attachments/${ATTACHMENT_ID}`,
    `blob:https://api.example.test/${ATTACHMENT_ID}`,
  ])(
    "does not promote an untrusted or temporary URL into a file preview",
    (url) => {
      const editor = new Editor({
        extensions,
        content: {
          type: "doc",
          content: [{ type: COMMENT_ATTACHMENT_NODE, attrs: { ...FILE, url } }],
        },
      });
      expect(splitCommentAttachments(editor.getHTML()).attachments).toEqual([]);
      editor.destroy();
    },
  );
});
