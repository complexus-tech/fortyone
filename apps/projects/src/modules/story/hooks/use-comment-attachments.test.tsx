/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { Editor } from "@tiptap/core";
import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { addAttachmentAction } from "../actions/add-attachment";
import { COMMENT_ATTACHMENT_NODE } from "../components/comment-attachments";
import { useCommentAttachments } from "./use-comment-attachments";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/hooks/subscription-features", () => ({
  useSubscriptionFeatures: () => ({ getLimit: () => "10MB" }),
}));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.test",
}));
jest.mock("../actions/add-attachment", () => ({
  addAttachmentAction: jest.fn(),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const mockUpload = jest.mocked(addAttachmentAction);
const ATTACHMENT_ID = "7d3a62e4-a7cc-4489-99b6-a986195525c5";
const FILE = new File(["image content"], "screenshot.png", {
  type: "image/png",
});
const RESPONSE = {
  data: {
    id: ATTACHMENT_ID,
    filename: FILE.name,
    url: "https://storage.example.test/image?token=temporary",
    mimeType: FILE.type,
    size: FILE.size,
    createdAt: "2026-10-16T12:00:00Z",
    uploadedBy: "author",
  },
  error: undefined,
};

const createEditor = () => {
  const insertContentAt = jest.fn().mockReturnValue(true);
  const editor = {
    isEditable: true,
    isDestroyed: false,
    state: { doc: { content: { size: 12 } } },
    commands: { insertContentAt },
  } as unknown as Editor;
  return { editor, insertContentAt };
};
const renderAttachments = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return renderHook(() => useCommentAttachments("story"), { wrapper });
};

beforeEach(() => {
  jest.clearAllMocks();
  mockUpload.mockReset();
  let sequence = 0;
  Object.defineProperty(globalThis.crypto, "randomUUID", {
    configurable: true,
    value: () => `upload-${++sequence}`,
  });
});

describe("comment attachment uploads", () => {
  it("blocks immediately while uploading and stores the durable authenticated URL on completion", async () => {
    let resolveUpload!: (value: typeof RESPONSE) => void;
    mockUpload.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveUpload = resolve;
        }),
    );
    const { editor, insertContentAt } = createEditor();
    const { result } = renderAttachments();
    act(() => {
      result.current.addFiles(editor, [FILE]);
      expect(result.current.hasPendingUploads()).toBe(true);
    });
    expect(result.current.hasBlockingUploads).toBe(true);
    expect(result.current.isUploading).toBe(true);
    expect(mockUpload).toHaveBeenCalledWith("story", FILE, "acme");
    await act(async () => {
      resolveUpload(RESPONSE);
    });
    expect(result.current.hasBlockingUploads).toBe(false);
    expect(insertContentAt).toHaveBeenCalledWith(12, [
      {
        type: COMMENT_ATTACHMENT_NODE,
        attrs: {
          ...RESPONSE.data,
          url: `https://api.example.test/workspaces/acme/stories/story/attachments/${ATTACHMENT_ID}`,
        },
      },
      { type: "paragraph" },
    ]);
  });

  it("keeps failed uploads blocking submission until a successful retry or removal", async () => {
    mockUpload.mockResolvedValueOnce({
      data: null,
      error: { message: "Offline" },
    });
    const { editor, insertContentAt } = createEditor();
    const { result } = renderAttachments();
    act(() => {
      result.current.addFiles(editor, [FILE]);
    });
    await waitFor(() => {
      expect(result.current.uploads[0]?.status).toBe("failed");
    });
    expect(result.current.uploads[0]?.error).toBe("Offline");
    expect(result.current.hasBlockingUploads).toBe(true);
    expect(result.current.isUploading).toBe(false);
    expect(insertContentAt).not.toHaveBeenCalled();
    mockUpload.mockResolvedValueOnce(RESPONSE);
    act(() => {
      result.current.retry(editor, result.current.uploads[0].id);
    });
    await waitFor(() => {
      expect(result.current.hasBlockingUploads).toBe(false);
    });
    expect(insertContentAt).toHaveBeenCalledTimes(1);

    mockUpload.mockResolvedValueOnce({
      data: null,
      error: { message: "Offline" },
    });
    act(() => {
      result.current.addFiles(editor, [FILE]);
    });
    await waitFor(() => {
      expect(result.current.uploads[0]?.status).toBe("failed");
    });
    act(() => {
      result.current.remove(result.current.uploads[0].id);
    });
    expect(result.current.hasBlockingUploads).toBe(false);
  });

  it("rejects unsupported, empty, oversized, and over-batch files before any upload", () => {
    const { editor } = createEditor();
    const { result } = renderAttachments();
    const oversized = new File(["x"], "large.png", { type: "image/png" });
    Object.defineProperty(oversized, "size", { value: 11 * 1024 * 1024 });
    act(() => {
      result.current.addFiles(editor, [
        new File(["x"], "program.exe", { type: "application/octet-stream" }),
        new File([], "empty.png", { type: "image/png" }),
        oversized,
      ]);
      result.current.addFiles(
        editor,
        Array.from({ length: 11 }, () => FILE),
      );
    });
    expect(mockUpload).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledTimes(4);
    expect(result.current.hasBlockingUploads).toBe(false);
  });

  it("does not mutate a destroyed editor when an upload finishes after navigation", async () => {
    let resolveUpload!: (value: typeof RESPONSE) => void;
    mockUpload.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveUpload = resolve;
        }),
    );
    const { editor, insertContentAt } = createEditor();
    const { result, unmount } = renderAttachments();
    act(() => {
      result.current.addFiles(editor, [FILE]);
    });
    unmount();
    await act(async () => {
      resolveUpload(RESPONSE);
    });
    expect(insertContentAt).not.toHaveBeenCalled();
  });
});
