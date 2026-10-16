/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { AttachmentTextPreview } from "./attachment-text-preview";
import { MAX_TEXT_PREVIEW_BYTES, readAttachmentText } from "./attachment-text";

const FILE = {
  id: "file",
  filename: "notes.txt",
  url: "https://api.example.test/workspaces/acme/stories/story/attachments/file",
  mimeType: "text/plain",
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
});

const mockTextResponse = (text: string) => {
  mockFetch.mockResolvedValueOnce({
    ok: true,
    json: async () => ({
      data: { url: "https://storage.example.test/signed" },
    }),
  });
  mockFetch.mockResolvedValueOnce({
    ok: true,
    body: new ReadableStream({
      start(controller) {
        controller.enqueue(new TextEncoder().encode(text));
        controller.close();
      },
    }),
  });
};
const renderPreview = () =>
  render(
    <QueryClientProvider
      client={
        new QueryClient({ defaultOptions: { queries: { retry: false } } })
      }
    >
      <AttachmentTextPreview file={FILE} />
    </QueryClientProvider>,
  );

describe("bounded attachment text preview", () => {
  it("authenticates resolution and renders file content as escaped text", async () => {
    const content = '<script>alert("x")</script><img src="private" />';
    mockTextResponse(content);
    const { container } = renderPreview();
    const preview = await screen.findByLabelText("notes.txt text preview");
    expect(preview.textContent).toBe(content);
    expect(container.querySelector("script")).toBeNull();
    expect(container.querySelector("img")).toBeNull();
    expect(mockFetch).toHaveBeenNthCalledWith(
      1,
      FILE.url,
      expect.objectContaining({
        credentials: "include",
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockFetch).toHaveBeenNthCalledWith(
      2,
      "https://storage.example.test/signed",
      expect.objectContaining({
        credentials: "omit",
        headers: { Range: "bytes=0-262143" },
      }),
    );
  });

  it("stops at the byte limit and cancels a storage response that ignores Range", async () => {
    const cancel = jest.fn();
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        data: { url: "https://storage.example.test/signed" },
      }),
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      body: new ReadableStream({
        start(controller) {
          controller.enqueue(
            new TextEncoder().encode("a".repeat(MAX_TEXT_PREVIEW_BYTES + 100)),
          );
        },
        cancel,
      }),
    });
    const result = await readAttachmentText(FILE, new AbortController().signal);
    expect(result.text.length).toBe(MAX_TEXT_PREVIEW_BYTES);
    expect(result.truncated).toBe(true);
    expect(cancel).toHaveBeenCalledTimes(1);
  });

  it("shows a recoverable error when an attachment is unavailable", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      json: async () => ({ error: { message: "Attachment unavailable" } }),
    });
    renderPreview();
    expect(
      await screen.findByText("Attachment unavailable"),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Retry preview" }),
    ).toBeInTheDocument();
  });
});
