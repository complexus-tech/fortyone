import { act, renderHook, waitFor } from "@testing-library/react";
import { HocuspocusProvider } from "@hocuspocus/provider";
import * as Y from "yjs";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { bindDesktopSession } from "@/shell/desktop/session-binding";
import { flushDesktopDrafts } from "@/shell/desktop/drafts";
import { createCollaborationSessionAction } from "./actions";
import { encodeCollaboration } from "./collaboration-draft";
import type { CollaborationDraft } from "./collaboration-draft";
import { useDesktopDocumentCollaboration } from "./use-desktop-document-collaboration";

jest.mock("@hocuspocus/provider", () => ({
  HocuspocusProvider: jest.fn(),
  WebSocketStatus: { Connected: "connected" },
}));
jest.mock("@/shell/desktop/bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("@/shell/desktop/session-binding", () => ({
  bindDesktopSession: jest.fn().mockResolvedValue(true),
}));
jest.mock("@/public-env", () => ({
  getPublicEnv: () => ({
    COLLABORATION_URL: "wss://collaboration.example.test",
  }),
}));
jest.mock("./actions", () => ({ createCollaborationSessionAction: jest.fn() }));

const scope = {
  accountId: "alice",
  workspaceId: "workspace",
  workspaceSlug: "acme",
};
const user = { id: "alice", name: "Alice" };
let stored: CollaborationDraft | null;
const read = jest.fn();
const write = jest.fn();
const exportRecovery = jest.fn();
type ProviderConfig = ConstructorParameters<typeof HocuspocusProvider>[0];
let config: ProviderConfig;
let provider: {
  isSynced: boolean;
  hasUnsyncedChanges: boolean;
  destroy: jest.Mock;
  disconnect: jest.Mock;
  setAwarenessField: jest.Mock;
};

beforeEach(() => {
  jest.clearAllMocks();
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: false,
  });
  const document = new Y.Doc();
  document.getText("title").insert(0, "Downloaded title");
  stored = {
    version: 1,
    kind: "collaboration",
    name: "doc:3",
    state: encodeCollaboration(document),
    dirty: false,
    canEdit: true,
    updatedAt: Date.now(),
  };
  document.destroy();
  read.mockImplementation(async () => stored);
  write.mockImplementation(
    async ({ payload }: { payload: CollaborationDraft }) => {
      stored = payload;
    },
  );
  exportRecovery.mockResolvedValue(false);
  jest.mocked(getDesktopBridge).mockReturnValue({
    draftRead: read,
    draftWrite: write,
    exportRecovery,
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);
  jest
    .mocked(createCollaborationSessionAction)
    .mockResolvedValue({ name: "doc:3", token: "ticket" });
  provider = {
    isSynced: false,
    hasUnsyncedChanges: false,
    destroy: jest.fn(),
    disconnect: jest.fn(),
    setAwarenessField: jest.fn(),
  };
  jest.mocked(HocuspocusProvider).mockImplementation((input) => {
    config = input;
    return provider as unknown as HocuspocusProvider;
  });
});
afterEach(async () => {
  await flushDesktopDrafts();
});
const mount = (epoch = 3) =>
  renderHook(() =>
    useDesktopDocumentCollaboration("doc", scope, user, {
      collaborationEpoch: epoch,
      canEdit: true,
    }),
  );

it("reopens a downloaded document offline, persists edits, and restores them after closing", async () => {
  const first = mount();
  await waitFor(() => {
    expect(first.result.current.ready).toBe(true);
  });
  expect(first.result.current.title).toBe("Downloaded title");
  expect(first.result.current.writable).toBe(true);
  expect(createCollaborationSessionAction).not.toHaveBeenCalled();
  act(() => {
    first.result.current.setTitle("Edited while offline");
  });
  await waitFor(() => {
    expect(first.result.current.localStatus).toContain("waiting to sync");
  });
  expect(write).toHaveBeenLastCalledWith(
    expect.objectContaining({
      accountId: "alice",
      workspaceId: "workspace",
      key: "collaboration:doc",
      payload: expect.objectContaining({ dirty: true }),
    }),
  );
  first.unmount();
  const reopened = mount();
  await waitFor(() => {
    expect(reopened.result.current.title).toBe("Edited while offline");
  });
  reopened.unmount();
});

it("does not report a local edit saved until the native write resolves", async () => {
  const view = mount();
  await waitFor(() => {
    expect(view.result.current.ready).toBe(true);
  });
  let finish!: () => void;
  write.mockImplementationOnce(
    () =>
      new Promise<void>((resolve) => {
        finish = resolve;
      }),
  );
  act(() => {
    view.result.current.setTitle("Awaiting disk");
  });
  await waitFor(() => {
    expect(write).toHaveBeenCalled();
  });
  expect(view.result.current.localStatus).toBe("Saving on this device…");
  await act(async () => {
    finish();
  });
  expect(view.result.current.localStatus).toContain("waiting to sync");
  view.unmount();
});

it("merges on reconnect and marks clean only after writable server acknowledgement", async () => {
  stored!.dirty = true;
  const view = mount();
  await waitFor(() => {
    expect(view.result.current.ready).toBe(true);
  });
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: true,
  });
  await act(async () => {
    window.dispatchEvent(new Event("online"));
  });
  await waitFor(() => {
    expect(HocuspocusProvider).toHaveBeenCalledTimes(1);
  });
  expect(config.document?.getText("title").toString()).toBe("Downloaded title");
  await act(async () => {
    config.onAuthenticated?.({ scope: "read-write" });
    provider.isSynced = true;
    provider.hasUnsyncedChanges = true;
    config.onSynced?.({ state: true });
  });
  expect(stored?.dirty).toBe(true);
  await act(async () => {
    provider.hasUnsyncedChanges = false;
    config.onUnsyncedChanges?.({ number: 0 });
  });
  await waitFor(() => {
    expect(stored?.dirty).toBe(false);
  });
  view.unmount();
});

it("retains divergent offline edits without uploading them to a restored document", async () => {
  stored!.dirty = true;
  const original = stored!.state;
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: true,
  });
  const view = mount(4);
  await waitFor(() => {
    expect(view.result.current.status).toBe("blocked");
  });
  expect(view.result.current.title).toBe("Downloaded title");
  expect(view.result.current.writable).toBe(false);
  expect(HocuspocusProvider).not.toHaveBeenCalled();
  expect(stored?.state).toBe(original);
  view.unmount();
});

it("rechecks the epoch on ticket renewal before sending existing Yjs state", async () => {
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: true,
  });
  const view = mount();
  await waitFor(() => {
    expect(HocuspocusProvider).toHaveBeenCalled();
  });
  jest
    .mocked(createCollaborationSessionAction)
    .mockResolvedValue({ name: "doc:4", token: "new-ticket" });
  await act(async () => {
    expect(typeof config.token).toBe("function");
    await expect((config.token as () => Promise<string>)()).rejects.toThrow(
      "version changed",
    );
  });
  expect(view.result.current.status).toBe("blocked");
  expect(provider.disconnect).toHaveBeenCalled();
  view.unmount();
});

it("keeps the editor locked if the native account/workspace lease cannot be verified", async () => {
  jest
    .mocked(bindDesktopSession)
    .mockRejectedValueOnce(new Error("Reconnect to renew access"));
  const view = mount();
  await waitFor(() => {
    expect(view.result.current.error).toContain("renew access");
  });
  expect(read).not.toHaveBeenCalled();
  expect(view.result.current.writable).toBe(false);
  expect(view.result.current.connection).toBeNull();
  view.unmount();
});

it("enables recovery replacement only after a native export actually saves the file", async () => {
  stored!.dirty = true;
  const view = mount(4);
  await waitFor(() => {
    expect(view.result.current.status).toBe("blocked");
  });
  await act(async () => {
    await view.result.current.exportRecovery();
  });
  expect(view.result.current.exported).toBe(false);
  await expect(view.result.current.openRestored()).rejects.toThrow(
    "Export the latest",
  );
  exportRecovery.mockResolvedValue(true);
  await act(async () => {
    await view.result.current.exportRecovery({
      contentHtml: "<p>Local work</p>",
      contentText: "Local work",
    });
  });
  expect(view.result.current.exported).toBe(true);
  expect(exportRecovery).toHaveBeenLastCalledWith(
    expect.objectContaining({
      accountId: "alice",
      workspaceId: "workspace",
      documentId: "doc",
      payload: expect.objectContaining({
        name: "doc:3",
        contentText: "Local work",
      }),
    }),
  );
  view.unmount();
});

it("retries authentication without discarding the retained document", async () => {
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: true,
  });
  const view = mount();
  await waitFor(() => {
    expect(HocuspocusProvider).toHaveBeenCalledTimes(1);
  });
  act(() => {
    config.onAuthenticationFailed?.({ reason: "expired" });
  });
  expect(view.result.current.writable).toBe(false);
  expect(view.result.current.status).toBe("blocked");
  await act(async () => {
    await view.result.current.retry();
  });
  expect(HocuspocusProvider).toHaveBeenCalledTimes(2);
  expect(view.result.current.title).toBe("Downloaded title");
  expect(view.result.current.writable).toBe(false);
  act(() => {
    config.onAuthenticated?.({ scope: "read-write" });
  });
  expect(view.result.current.writable).toBe(true);
  view.unmount();
});

it("freezes offline editing on an explicit access denial and permits a later authorized retry", async () => {
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: true,
  });
  jest
    .mocked(createCollaborationSessionAction)
    .mockRejectedValueOnce(
      Object.assign(new Error("Forbidden"), { status: 403 }),
    );
  const view = mount();
  await waitFor(() => {
    expect(view.result.current.status).toBe("blocked");
  });
  expect(view.result.current.writable).toBe(false);
  expect(HocuspocusProvider).not.toHaveBeenCalled();
  await act(async () => {
    await view.result.current.retry();
  });
  expect(HocuspocusProvider).toHaveBeenCalledTimes(1);
  view.unmount();
});

it("keeps a failed local save actionable after the connection recovers", async () => {
  const view = mount();
  await waitFor(() => {
    expect(view.result.current.ready).toBe(true);
  });
  write.mockRejectedValueOnce(new Error("Disk is full"));
  act(() => {
    view.result.current.setTitle("Unsaved local title");
  });
  await waitFor(() => {
    expect(view.result.current.storageError).toBe("Disk is full");
  });
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: true,
  });
  await act(async () => {
    window.dispatchEvent(new Event("online"));
  });
  expect(view.result.current.error).toBeNull();
  expect(view.result.current.storageError).toBe("Disk is full");
  act(() => {
    view.result.current.setTitle("Successfully saved local title");
  });
  await waitFor(() => {
    expect(view.result.current.storageError).toBeNull();
  });
  view.unmount();
});
