import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as Y from "yjs";
import type { CollaborationDraft } from "../collaboration";
import type { Document } from "../types";
import { encodeState, replaceTitle } from "../collaboration";
import { native } from "../native";
import { useCollaboration } from "./use-collaboration";
import { flushDrafts } from "../drafts";

const providers = vi.hoisted(
  () =>
    [] as Array<{
      configuration: Record<string, any>;
      disconnect: ReturnType<typeof vi.fn>;
      destroy: ReturnType<typeof vi.fn>;
    }>,
);
vi.mock("@hocuspocus/provider", () => ({
  WebSocketStatus: { Connected: "connected" },
  HocuspocusProvider: class {
    hasUnsyncedChanges = false;
    isSynced = true;
    disconnect = vi.fn();
    destroy = vi.fn();
    constructor(public configuration: Record<string, any>) {
      providers.push(this);
    }
  },
}));
vi.mock("../native", () => ({
  getBoundAccountId: () => "user",
  bindNativeAccount: vi.fn(),
  native: {
    forAccount: () => native,
    readDraft: vi.fn(),
    saveDraft: vi.fn(),
    collaborationSession: vi.fn(),
    exportDocument: vi.fn(),
    deleteDraft: vi.fn(),
  },
  errorMessage: (error: unknown) =>
    error instanceof Error ? error.message : String(error),
}));
const server = new Y.Doc();
replaceTitle(server.getText("title"), "Original");
const snapshot: Document = {
  id: "doc",
  title: "Original",
  contentHtml: "<p></p>",
  contentText: "",
  revision: 1,
  collaborationEpoch: 1,
  collaborationState: encodeState(server),
  canEdit: true,
  updatedAt: "2026-10-14",
  collaborative: true,
};
const stored: CollaborationDraft = {
  name: "doc:1",
  epoch: 1,
  state: encodeState(server),
  dirty: true,
  title: "Original",
  contentHtml: "",
  contentText: "",
};

beforeEach(() => {
  vi.clearAllMocks();
  providers.length = 0;
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: false,
  });
  vi.mocked(native.readDraft).mockResolvedValue(null);
  vi.mocked(native.saveDraft).mockResolvedValue();
  vi.mocked(native.collaborationSession).mockResolvedValue({
    name: "doc:1",
    token: "ticket",
    url: "wss://collaboration.example.com",
  });
});

describe("durable collaborative editor lifecycle", () => {
  it("keeps editing offline and only reports local save after native persistence", async () => {
    const { result } = renderHook(() =>
      useCollaboration(snapshot, "workspace"),
    );
    await waitFor(() => expect(result.current.ready).toBe(true));
    expect(native.collaborationSession).not.toHaveBeenCalled();
    act(() => result.current.setTitle("An offline change"));
    await act(async () => {
      await flushDrafts();
    });
    expect(result.current.title).toBe("An offline change");
    expect(result.current.status).toBe(
      "Saved on this device · waiting to sync",
    );
    expect(native.saveDraft).toHaveBeenLastCalledWith(
      "workspace",
      "doc",
      "collaboration",
      expect.objectContaining({
        dirty: true,
        title: "An offline change",
        name: "doc:1",
      }),
    );
  });
  it("does not mark local changes synchronized when edit permission was revoked", async () => {
    Object.defineProperty(navigator, "onLine", {
      configurable: true,
      value: true,
    });
    vi.mocked(native.readDraft).mockResolvedValue({ payload: stored });
    const { result } = renderHook(() =>
      useCollaboration(snapshot, "workspace"),
    );
    await waitFor(() => expect(providers).toHaveLength(1));
    await act(async () => {
      providers[0].configuration.onAuthenticated({ scope: "read-only" });
      providers[0].configuration.onSynced({ state: true });
      await flushDrafts();
    });
    expect(result.current.blocked).toBe(true);
    expect(result.current.writable).toBe(false);
    expect(native.saveDraft).toHaveBeenLastCalledWith(
      "workspace",
      "doc",
      "collaboration",
      expect.objectContaining({ dirty: true }),
    );
  });
  it("blocks reconnect before sending old edits into a restored document epoch", async () => {
    Object.defineProperty(navigator, "onLine", {
      configurable: true,
      value: true,
    });
    vi.mocked(native.readDraft).mockResolvedValue({ payload: stored });
    const { result } = renderHook(() =>
      useCollaboration(snapshot, "workspace"),
    );
    await waitFor(() => expect(providers).toHaveLength(1));
    vi.mocked(native.collaborationSession).mockResolvedValue({
      name: "doc:2",
      token: "new-ticket",
      url: "wss://collaboration.example.com",
    });
    await act(async () => {
      await expect(providers[0].configuration.token()).rejects.toThrow(
        "restored",
      );
    });
    expect(providers[0].disconnect).toHaveBeenCalled();
    expect(result.current.blocked).toBe(true);
    expect(result.current.connection?.name).toBe("doc:1");
  });
});
