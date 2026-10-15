import { beforeEach, describe, expect, it, vi } from "vitest";
import * as Y from "yjs";
import { uploadCollaboration } from "./collaboration-upload";

type Callbacks = {
  onAuthenticated: (event: { scope: string }) => void;
  onSynced: () => void;
  onUnsyncedChanges: () => void;
};
const providers = vi.hoisted(
  () =>
    [] as Array<{
      options: Callbacks;
      isSynced: boolean;
      hasUnsyncedChanges: boolean;
      destroy: ReturnType<typeof vi.fn>;
    }>,
);
vi.mock("@hocuspocus/provider", () => ({
  HocuspocusProvider: class {
    isSynced = true;
    hasUnsyncedChanges = true;
    destroy = vi.fn();
    constructor(public options: Callbacks) {
      providers.push(this);
    }
  },
}));
const ticket = {
  name: "doc:1",
  token: "ticket",
  url: "wss://collaboration.example",
};
beforeEach(() => {
  providers.length = 0;
});

describe("background document acknowledgment", () => {
  it("waits for both write permission and acknowledgment of every local update", async () => {
    const document = new Y.Doc();
    const finished = vi.fn();
    const upload = uploadCollaboration(
      ticket,
      document,
      new AbortController().signal,
    ).then(finished);
    const provider = providers[0]!;
    provider.options.onSynced();
    provider.options.onAuthenticated({ scope: "read-write" });
    await Promise.resolve();
    expect(finished).not.toHaveBeenCalled();
    provider.hasUnsyncedChanges = false;
    provider.options.onUnsyncedChanges();
    await upload;
    expect(finished).toHaveBeenCalledWith(true);
    expect(provider.destroy).toHaveBeenCalledOnce();
    document.destroy();
  });
  it("preserves dirty work when permission becomes read-only", async () => {
    const document = new Y.Doc();
    const upload = uploadCollaboration(
      ticket,
      document,
      new AbortController().signal,
    );
    const provider = providers[0]!;
    provider.hasUnsyncedChanges = false;
    provider.options.onSynced();
    provider.options.onAuthenticated({ scope: "read-only" });
    expect(await upload).toBe(false);
    expect(provider.destroy).toHaveBeenCalledOnce();
    document.destroy();
  });
  it("closes the provider without acknowledging work when its owner exits", async () => {
    const document = new Y.Doc();
    const controller = new AbortController();
    const upload = uploadCollaboration(ticket, document, controller.signal);
    controller.abort();
    expect(await upload).toBe(false);
    expect(providers[0]!.destroy).toHaveBeenCalledOnce();
    document.destroy();
  });
});
