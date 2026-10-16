/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getCommentFileType } from "./comment-file-type";

describe("comment file types", () => {
  it.each([
    ["application/msword", "word"],
    [
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      "word",
    ],
    ["application/vnd.ms-excel", "excel"],
    [
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
      "excel",
    ],
    ["application/vnd.ms-powerpoint", "powerpoint"],
    [
      "application/vnd.openxmlformats-officedocument.presentationml.presentation",
      "powerpoint",
    ],
    ["application/pdf", "pdf"],
    ["text/csv", "csv"],
    ["text/plain; charset=utf-8", "text"],
    ["video/mp4", "video"],
    ["image/png", "image"],
  ])("identifies %s from its server content type", (mimeType, expected) => {
    expect(getCommentFileType(mimeType)).toBe(expected);
  });

  it("prefers a recognized content type over a conflicting extension", () => {
    expect(getCommentFileType("application/pdf", "report.xlsx")).toBe("pdf");
    expect(getCommentFileType("text/plain", "report.docx")).toBe("text");
  });

  it("handles MIME casing and older generic Office upload metadata", () => {
    expect(getCommentFileType(" Application/PDF ")).toBe("pdf");
    expect(getCommentFileType("application/zip", "Quarterly.DOCX")).toBe(
      "word",
    );
    expect(getCommentFileType("application/octet-stream", "Budget.XLS")).toBe(
      "excel",
    );
    expect(getCommentFileType("", "Update.PPTX")).toBe("powerpoint");
    expect(getCommentFileType("text/plain; charset=utf-8", "rows.csv")).toBe(
      "csv",
    );
  });

  it("keeps unsupported archives and unknown files generic", () => {
    expect(getCommentFileType("application/zip", "backup.zip")).toBe("file");
    expect(getCommentFileType("application/x-rar", "backup.rar")).toBe("file");
    expect(getCommentFileType("application/octet-stream", "unknown")).toBe(
      "file",
    );
    expect(getCommentFileType("application/octet-stream", "docx")).toBe("file");
    expect(getCommentFileType("application/x-custom", "custom.pdf")).toBe(
      "file",
    );
  });
});
