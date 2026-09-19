/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  CHAT_ATTACHMENT_ACCEPT,
  getChatAttachmentMediaType,
  WORD_DOCUMENT_MIME_TYPES,
} from "./chat-attachment-types";

describe("Maya chat attachment types", () => {
  it("accepts legacy and modern Word document extensions", () => {
    expect(CHAT_ATTACHMENT_ACCEPT[WORD_DOCUMENT_MIME_TYPES.doc]).toEqual([
      ".doc",
    ]);
    expect(CHAT_ATTACHMENT_ACCEPT[WORD_DOCUMENT_MIME_TYPES.docx]).toEqual([
      ".docx",
    ]);
  });

  it("normalizes Word MIME types from the filename", () => {
    const doc = new File(["legacy"], "BRIEF.DOC", {
      type: "application/octet-stream",
    });
    const docx = new File(["modern"], "brief.docx", {
      type: "application/zip",
    });

    expect(getChatAttachmentMediaType(doc)).toBe(WORD_DOCUMENT_MIME_TYPES.doc);
    expect(getChatAttachmentMediaType(docx)).toBe(
      WORD_DOCUMENT_MIME_TYPES.docx,
    );
  });

  it("preserves the reported MIME type for other supported files", () => {
    const pdf = new File(["pdf"], "brief.pdf", { type: "application/pdf" });

    expect(getChatAttachmentMediaType(pdf)).toBe("application/pdf");
  });
});
