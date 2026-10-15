import { describe, expect, it, vi } from "vitest";
import { flushDrafts, persistDraft, retryFailedDrafts } from "./drafts";
import { native } from "./native";
vi.mock("./native", () => ({
  getBoundAccountId: () => "user",
  native: { forAccount: () => native, saveDraft: vi.fn() },
}));

describe("durable draft ordering", () => {
  it("serializes writes to prevent an earlier value replacing the newest input", async () => {
    let resolveFirst!: () => void;
    vi.mocked(native.saveDraft)
      .mockImplementationOnce(
        () =>
          new Promise<void>((resolve) => {
            resolveFirst = resolve;
          }),
      )
      .mockResolvedValueOnce();
    const first = persistDraft("workspace", "ordered", "story", { title: "A" });
    const second = persistDraft("workspace", "ordered", "story", {
      title: "AB",
    });
    await Promise.resolve();
    await Promise.resolve();
    expect(native.saveDraft).toHaveBeenCalledTimes(1);
    resolveFirst();
    await Promise.all([first, second]);
    expect(native.saveDraft).toHaveBeenLastCalledWith(
      "workspace",
      "ordered",
      "story",
      { title: "AB" },
    );
    await expect(flushDrafts()).resolves.toBeUndefined();
  });
  it("blocks leaving a workspace after a disk failure until the draft is saved", async () => {
    vi.mocked(native.saveDraft).mockRejectedValueOnce(new Error("Disk full"));
    await expect(
      persistDraft("workspace", "disk-failure", "document", {
        title: "Important",
      }),
    ).rejects.toThrow("Disk full");
    await expect(flushDrafts()).rejects.toThrow("A draft could not be saved");
    vi.mocked(native.saveDraft).mockResolvedValueOnce();
    await persistDraft("workspace", "disk-failure", "document", {
      title: "Important",
    });
    await expect(flushDrafts()).resolves.toBeUndefined();
  });
});

describe("retained draft recovery before browser reconnection", () => {
  it("retains a failed collaboration checkpoint as pending recovery until access can be renewed", async () => {
    vi.mocked(native.saveDraft).mockRejectedValueOnce(
      new Error("Reconnect to renew access"),
    );
    await expect(
      persistDraft(
        "workspace",
        "collab",
        "collaboration",
        { name: "doc:1", state: "encoded-state", dirty: false },
        "account-a",
      ),
    ).rejects.toThrow();
    vi.mocked(native.saveDraft).mockResolvedValueOnce();
    await retryFailedDrafts("account-a");
    expect(native.saveDraft).toHaveBeenLastCalledWith(
      "workspace",
      "collab",
      "collaboration",
      { name: "doc:1", state: "encoded-state", dirty: true },
    );
  });
  it("retains a newer failed keystroke that arrives while its earlier retry is in progress", async () => {
    vi.mocked(native.saveDraft).mockRejectedValueOnce(
      new Error("Lease expired"),
    );
    await expect(
      persistDraft(
        "workspace",
        "during-retry",
        "story",
        { title: "Before retry" },
        "account-a",
      ),
    ).rejects.toThrow();
    let finish!: () => void;
    vi.mocked(native.saveDraft).mockImplementationOnce(
      () =>
        new Promise<void>((resolve) => {
          finish = resolve;
        }),
    );
    const reconnect = retryFailedDrafts("account-a");
    await Promise.resolve();
    await Promise.resolve();
    await Promise.resolve();
    vi.mocked(native.saveDraft).mockRejectedValueOnce(new Error("Disk full"));
    const newest = persistDraft(
      "workspace",
      "during-retry",
      "story",
      { title: "Typed during retry" },
      "account-a",
    );
    const newestRejected = expect(newest).rejects.toThrow("Disk full");
    const reconnectRejected = expect(reconnect).rejects.toThrow(
      "A draft could not be saved",
    );
    finish();
    await Promise.all([newestRejected, reconnectRejected]);
    vi.mocked(native.saveDraft).mockResolvedValueOnce();
    await retryFailedDrafts("account-a");
    expect(native.saveDraft).toHaveBeenLastCalledWith(
      "workspace",
      "during-retry",
      "story",
      { title: "Typed during retry" },
    );
  });

  it("retries only the newest failed payload, scoped to its initiating account", async () => {
    vi.mocked(native.saveDraft).mockRejectedValueOnce(
      new Error("Reconnect to renew access"),
    );
    await expect(
      persistDraft(
        "workspace",
        "recover",
        "story",
        { title: "Older" },
        "account-a",
      ),
    ).rejects.toThrow();
    vi.mocked(native.saveDraft).mockRejectedValueOnce(
      new Error("Reconnect to renew access"),
    );
    await expect(
      persistDraft(
        "workspace",
        "recover",
        "story",
        { title: "Latest retained edit" },
        "account-a",
      ),
    ).rejects.toThrow();
    const before = vi.mocked(native.saveDraft).mock.calls.length;
    await expect(retryFailedDrafts("account-b")).rejects.toThrow(
      "A draft could not be saved",
    );
    expect(native.saveDraft).toHaveBeenCalledTimes(before);
    vi.mocked(native.saveDraft).mockResolvedValueOnce();
    await retryFailedDrafts("account-a");
    expect(native.saveDraft).toHaveBeenCalledTimes(before + 1);
    expect(native.saveDraft).toHaveBeenLastCalledWith(
      "workspace",
      "recover",
      "story",
      { title: "Latest retained edit" },
    );
    await expect(flushDrafts()).resolves.toBeUndefined();
  });
  it("preserves failed payloads through another failed retry until durable storage recovers", async () => {
    vi.mocked(native.saveDraft).mockRejectedValueOnce(new Error("Disk full"));
    await expect(
      persistDraft(
        "workspace",
        "retry-disk",
        "document",
        { title: "Keep this draft" },
        "account-a",
      ),
    ).rejects.toThrow();
    vi.mocked(native.saveDraft).mockRejectedValueOnce(new Error("Disk full"));
    await expect(retryFailedDrafts("account-a")).rejects.toThrow(
      "A draft could not be saved",
    );
    vi.mocked(native.saveDraft).mockResolvedValueOnce();
    await retryFailedDrafts("account-a");
    expect(native.saveDraft).toHaveBeenLastCalledWith(
      "workspace",
      "retry-disk",
      "document",
      { title: "Keep this draft" },
    );
  });
  it("does not replay an older failure after a newer edit is already being saved", async () => {
    vi.mocked(native.saveDraft).mockRejectedValueOnce(
      new Error("Lease expired"),
    );
    await expect(
      persistDraft(
        "workspace",
        "race",
        "story",
        { title: "Old failure" },
        "account-a",
      ),
    ).rejects.toThrow();
    let finish!: () => void;
    vi.mocked(native.saveDraft).mockImplementationOnce(
      () =>
        new Promise<void>((resolve) => {
          finish = resolve;
        }),
    );
    const latest = persistDraft(
      "workspace",
      "race",
      "story",
      { title: "New input" },
      "account-a",
    );
    const reconnect = retryFailedDrafts("account-a");
    await Promise.resolve();
    await Promise.resolve();
    const before = vi.mocked(native.saveDraft).mock.calls.length;
    finish();
    await Promise.all([latest, reconnect]);
    expect(native.saveDraft).toHaveBeenCalledTimes(before);
    expect(native.saveDraft).toHaveBeenLastCalledWith(
      "workspace",
      "race",
      "story",
      { title: "New input" },
    );
  });
});
