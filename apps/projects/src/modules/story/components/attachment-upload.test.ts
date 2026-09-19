/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { FileRejection } from "react-dropzone";
import {
  getAttachmentRejectionMessage,
  PAID_ATTACHMENT_SIZE_LIMIT,
  STORY_ATTACHMENT_ACCEPT,
  uploadAttachmentsConcurrently,
} from "./attachment-upload";

describe("attachment uploads", () => {
  it("starts every accepted upload without waiting for the previous file", async () => {
    const files = [
      new File(["one"], "one.pdf", { type: "application/pdf" }),
      new File(["two"], "two.pdf", { type: "application/pdf" }),
      new File(["three"], "three.pdf", { type: "application/pdf" }),
    ];
    const resolvers: (() => void)[] = [];
    const upload = jest.fn(
      () =>
        new Promise<void>((resolve) => {
          resolvers.push(resolve);
        }),
    );

    const batch = uploadAttachmentsConcurrently(files, upload);

    expect(upload).toHaveBeenCalledTimes(3);
    resolvers.forEach((resolve) => {
      resolve();
    });
    await expect(batch).resolves.toHaveLength(3);
  });

  it("returns a specific per-file size error", () => {
    const file = new File(["large"], "large-photo.jpg", {
      type: "image/jpeg",
    });
    Object.defineProperty(file, "size", { value: 26 * 1024 * 1024 });
    const rejection = {
      file,
      errors: [{ code: "file-too-large", message: "File is too large" }],
    } satisfies FileRejection;

    expect(
      getAttachmentRejectionMessage(rejection, PAID_ATTACHMENT_SIZE_LIMIT),
    ).toBe("large-photo.jpg is 26 MB. The maximum file size is 25 MB.");
  });

  it("accepts the supported document formats", () => {
    expect(STORY_ATTACHMENT_ACCEPT).toMatchObject({
      "application/msword": [".doc"],
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
        [".docx"],
      "application/vnd.ms-excel": [".xls"],
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [
        ".xlsx",
      ],
      "application/vnd.ms-powerpoint": [".ppt"],
      "application/vnd.openxmlformats-officedocument.presentationml.presentation":
        [".pptx"],
      "text/plain": [".txt"],
      "text/csv": [".csv"],
    });
  });

  it("describes every supported document family in invalid-type errors", () => {
    const file = new File(["archive"], "archive.zip", {
      type: "application/zip",
    });
    const rejection = {
      file,
      errors: [{ code: "file-invalid-type", message: "Invalid type" }],
    } satisfies FileRejection;

    expect(
      getAttachmentRejectionMessage(rejection, PAID_ATTACHMENT_SIZE_LIMIT),
    ).toContain("Word, Excel, PowerPoint, text, or CSV");
  });
});
