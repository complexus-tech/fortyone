import { act, renderHook } from "@testing-library/react";
import { uploadAttachment, deleteAttachment } from "./api";
import { useComposerAttachments } from "./composer-attachments";
import type { ChatDraft } from "./chat-rich-text";
import type { ChatAttachment } from "./types";

let mockConversationId = "room-1";
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
}));
jest.mock("./chat-context", () => ({
  useChatIdentity: () => ({ conversationId: mockConversationId }),
}));
jest.mock("./api", () => ({
  uploadAttachment: jest.fn(),
  deleteAttachment: jest.fn(),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
const upload = jest.mocked(uploadAttachment);
const remove = jest.mocked(deleteAttachment);
const uploaded: ChatAttachment = {
  id: "voice-1",
  fileName: "Voice note.webm",
  mediaType: "audio/webm",
  sizeBytes: 10,
  url: "",
  kind: "voice-note",
  durationMs: 2500,
};
const file = new File(["voice"], "Voice note.webm", { type: "audio/webm" });
const initial: ChatDraft = { text: "Original draft", content: null };

beforeEach(() => {
  jest.resetAllMocks();
  mockConversationId = "room-1";
  remove.mockResolvedValue({ data: null });
});
describe("composer upload completion", () => {
  it("merges voice uploads into the latest text and attachment draft", async () => {
    let finish!: (attachment: ChatAttachment) => void;
    upload.mockReturnValue(
      new Promise((resolve) => {
        finish = resolve;
      }),
    );
    const onDraftChange = jest.fn();
    const { result, rerender } = renderHook(
      ({ draft }) => useComposerAttachments(draft, onDraftChange, false),
      { initialProps: { draft: initial } },
    );
    let pending!: Promise<boolean>;
    act(() => {
      pending = result.current.uploadFiles([file], {
        kind: "voice-note",
        durationMs: 2500,
      });
    });
    const existing: ChatAttachment = {
      ...uploaded,
      id: "existing",
      kind: undefined,
    };
    const latest = {
      text: "Changed while uploading",
      content: { type: "doc", content: [] },
      attachments: [existing],
    };
    rerender({ draft: latest });
    await act(async () => {
      finish(uploaded);
      expect(await pending).toBe(true);
    });
    expect(onDraftChange).toHaveBeenCalledWith({
      ...latest,
      attachments: [existing, uploaded],
    });
    expect(upload).toHaveBeenCalledWith(
      { workspaceSlug: "workspace" },
      "room-1",
      file,
      { kind: "voice-note", durationMs: 2500 },
    );
  });
  it("blocks duplicate upload clicks and returns failure without discarding the draft", async () => {
    let reject!: (error: Error) => void;
    upload.mockReturnValue(
      new Promise((_resolve, fail) => {
        reject = fail;
      }),
    );
    const onDraftChange = jest.fn();
    const { result } = renderHook(() =>
      useComposerAttachments(initial, onDraftChange, false),
    );
    let pending!: Promise<boolean>;
    await act(async () => {
      pending = result.current.uploadFiles([file]);
      expect(await result.current.uploadFiles([file])).toBe(false);
    });
    expect(upload).toHaveBeenCalledTimes(1);
    await act(async () => {
      reject(new Error("Network offline"));
      expect(await pending).toBe(false);
    });
    expect(onDraftChange).not.toHaveBeenCalled();
    expect(result.current.uploading).toBe(false);
  });
  it("does not attach a late upload to another conversation", async () => {
    let finish!: (attachment: ChatAttachment) => void;
    upload.mockReturnValue(
      new Promise((resolve) => {
        finish = resolve;
      }),
    );
    const onDraftChange = jest.fn();
    const { result, rerender } = renderHook(() =>
      useComposerAttachments(initial, onDraftChange, false),
    );
    let pending!: Promise<boolean>;
    act(() => {
      pending = result.current.uploadFiles([file]);
    });
    mockConversationId = "room-2";
    rerender();
    await act(async () => {
      finish(uploaded);
      expect(await pending).toBe(false);
    });
    expect(onDraftChange).not.toHaveBeenCalled();
    expect(remove).toHaveBeenCalledWith(
      { workspaceSlug: "workspace" },
      "room-1",
      "voice-1",
    );
  });
  it("cleans up completed uploads after the composer unmounts", async () => {
    let finish!: (attachment: ChatAttachment) => void;
    upload.mockReturnValue(
      new Promise((resolve) => {
        finish = resolve;
      }),
    );
    const onDraftChange = jest.fn();
    const { result, unmount } = renderHook(() =>
      useComposerAttachments(initial, onDraftChange, false),
    );
    let pending!: Promise<boolean>;
    act(() => {
      pending = result.current.uploadFiles([file]);
    });
    unmount();
    await act(async () => {
      finish(uploaded);
      expect(await pending).toBe(false);
    });
    expect(onDraftChange).not.toHaveBeenCalled();
    expect(remove).toHaveBeenCalledTimes(1);
  });
});
