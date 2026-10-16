/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { downloadCommentAttachment } from "./comment-attachment-download";

const FILE = {
  id: "file",
  filename: "review.pdf",
  url: "https://api.example.test/workspaces/acme/stories/story/attachments/file",
  mimeType: "application/pdf",
  size: 128,
  uploadedBy: "author",
  createdAt: "2026-10-16T12:00:00Z",
};
const originalFetch = globalThis.fetch;
const mockFetch = jest.fn();

beforeEach(() => {
  mockFetch.mockReset();
  globalThis.fetch = mockFetch;
});
afterEach(() => {
  globalThis.fetch = originalFetch;
  jest.restoreAllMocks();
});

describe("comment attachment download", () => {
  it("authenticates URL resolution, omits credentials at storage, and downloads using the filename", async () => {
    const signedURL = "https://storage.example.test/file?signature=fresh";
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: { url: signedURL } }),
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      blob: async () => new Blob(["document"]),
    });
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: jest.fn(() => "blob:download"),
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: jest.fn(),
    });
    let downloadedFilename = "";
    jest
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(function captureDownload(this: HTMLAnchorElement) {
        downloadedFilename = this.download;
      });
    await downloadCommentAttachment(FILE);
    expect(mockFetch).toHaveBeenNthCalledWith(1, FILE.url, {
      credentials: "include",
      headers: { Accept: "application/json" },
      cache: "no-store",
    });
    expect(mockFetch).toHaveBeenNthCalledWith(2, signedURL, {
      credentials: "omit",
    });
    expect(downloadedFilename).toBe(FILE.filename);
  });

  it("surfaces an unavailable file without requesting storage", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      json: async () => ({ error: { message: "Attachment not found" } }),
    });
    await expect(downloadCommentAttachment(FILE)).rejects.toThrow(
      "Attachment not found",
    );
    expect(mockFetch).toHaveBeenCalledTimes(1);
  });
});
