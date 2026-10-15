import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import * as Y from "yjs";
import { native } from "../native";
import { encodeState, type CollaborationDraft } from "../collaboration";
import { uploadCollaboration } from "../collaboration-upload";
import { registerCollaborationEditor } from "../collaboration-registry";
import { useCollaborativeSync } from "./use-collaborative-sync";

vi.mock("../native", () => ({
  native: {
    forAccount: () => native,
    pendingCollaborations: vi.fn(),
    collaborationSession: vi.fn(),
    acknowledgeCollaboration: vi.fn(),
    refreshResource: vi.fn(),
    saveDraft: vi.fn(),
    listen: vi.fn().mockResolvedValue(() => undefined),
  },
  errorMessage: String,
}));
vi.mock("../drafts", () => ({
  flushDrafts: vi.fn().mockResolvedValue(undefined),
}));
vi.mock("../collaboration-upload", () => ({ uploadCollaboration: vi.fn() }));

describe("background collaboration durability", () => {
  it("does not overwrite a newer edit that was opened, changed, and closed during upload", async () => {
    Object.defineProperty(navigator, "onLine", {
      configurable: true,
      value: true,
    });
    const original = new Y.Doc();
    original.getText("title").insert(0, "Original");
    const first: CollaborationDraft = {
      name: "doc:1",
      epoch: 1,
      state: encodeState(original),
      title: "Original",
      contentHtml: "",
      contentText: "",
      dirty: true,
    };
    let durable = first;
    vi.mocked(native.pendingCollaborations).mockImplementation(async () => [
      { workspaceId: "workspace", documentId: "doc", payload: durable },
    ]);
    vi.mocked(native.collaborationSession).mockResolvedValue({
      name: "doc:1",
      token: "ticket",
      url: "wss://example.com",
    });
    vi.mocked(native.acknowledgeCollaboration).mockImplementation(
      async (_workspace, _document, expectedState, expectedName, next) => {
        if (durable.state !== expectedState || durable.name !== expectedName)
          return false;
        durable = next;
        return true;
      },
    );
    let acknowledge!: (success: boolean) => void;
    vi.mocked(uploadCollaboration).mockReturnValue(
      new Promise((resolve) => {
        acknowledge = resolve;
      }),
    );
    const { result } = renderHook(() => useCollaborativeSync("account"));
    await waitFor(() => expect(uploadCollaboration).toHaveBeenCalledOnce());
    const close = registerCollaborationEditor("workspace", "doc");
    original.getText("title").insert(8, " updated while uploading");
    durable = {
      ...first,
      state: encodeState(original),
      title: original.getText("title").toString(),
      dirty: true,
    };
    const latest = durable;
    close();
    await act(async () => {
      acknowledge(true);
    });
    await waitFor(() =>
      expect(native.acknowledgeCollaboration).toHaveBeenCalledOnce(),
    );
    expect(native.acknowledgeCollaboration).toHaveBeenCalledWith(
      "workspace",
      "doc",
      first.state,
      first.name,
      expect.objectContaining({ dirty: false }),
    );
    expect(durable).toBe(latest);
    expect(durable.dirty).toBe(true);
    expect(native.saveDraft).not.toHaveBeenCalled();
    expect(result.current.pendingCount).toBe(1);
    original.destroy();
  });
});
