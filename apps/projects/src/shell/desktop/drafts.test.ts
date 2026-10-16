import { getDesktopBridge } from "./bridge";
import {
  flushDesktopDrafts,
  removeDesktopDraft,
  retryDesktopDrafts,
  writeDesktopDraft,
} from "./drafts";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
const scope = {
  accountId: "account-a",
  workspaceId: "workspace",
  workspaceSlug: "acme",
};
const write = jest.fn();
beforeEach(() => {
  jest.clearAllMocks();
  write.mockResolvedValue(undefined);
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue({ draftWrite: write } as unknown as NonNullable<
      ReturnType<typeof getDesktopBridge>
    >);
});

it("retains the latest failed payload and retries only its initiating account", async () => {
  write.mockRejectedValueOnce(new Error("disk unavailable"));
  await expect(
    writeDesktopDraft(scope, "story:one", { title: "Latest local title" }),
  ).rejects.toThrow();
  await expect(flushDesktopDrafts()).rejects.toThrow("could not be saved");
  await expect(
    retryDesktopDrafts({ ...scope, accountId: "account-b" }),
  ).rejects.toThrow();
  expect(write).toHaveBeenCalledTimes(1);
  await retryDesktopDrafts(scope);
  expect(write).toHaveBeenLastCalledWith({
    accountId: "account-a",
    workspaceId: "workspace",
    key: "story:one",
    payload: { title: "Latest local title" },
  });
  await expect(flushDesktopDrafts()).resolves.toBeUndefined();
});

it("serializes a newer edit after an in-flight retry so the retry cannot overwrite it", async () => {
  write.mockRejectedValueOnce(new Error("disk unavailable"));
  await expect(
    writeDesktopDraft(scope, "story:two", { title: "First" }),
  ).rejects.toThrow();
  let finish!: () => void;
  write.mockImplementationOnce(
    () =>
      new Promise<void>((resolve) => {
        finish = resolve;
      }),
  );
  const retry = retryDesktopDrafts(scope);
  await new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
  const newer = writeDesktopDraft(scope, "story:two", { title: "Newest" });
  finish();
  await Promise.all([retry, newer]);
  expect(write.mock.calls.map(([request]) => request.payload.title)).toEqual([
    "First",
    "First",
    "Newest",
  ]);
  await expect(flushDesktopDrafts()).resolves.toBeUndefined();
});

it("retries failed collaboration snapshots as dirty recovery content after lease expiry", async () => {
  write.mockRejectedValueOnce(new Error("Reconnect to renew access"));
  await expect(
    writeDesktopDraft(scope, "collaboration:doc", {
      kind: "collaboration",
      state: "saved-yjs",
      dirty: false,
    }),
  ).rejects.toThrow();
  await retryDesktopDrafts(scope);
  expect(write).toHaveBeenLastCalledWith(
    expect.objectContaining({
      payload: { kind: "collaboration", state: "saved-yjs", dirty: true },
    }),
  );
});

it("preserves a newer native draft when an earlier submitted comment completes", async () => {
  const remove = jest.fn();
  jest.mocked(getDesktopBridge).mockReturnValue({
    draftWrite: write,
    draftRead: jest.fn().mockResolvedValue({ html: "Newer draft" }),
    draftDelete: remove,
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);

  const removed = await removeDesktopDraft(
    scope,
    "comment:story",
    (payload) => (payload as { html: string }).html === "Submitted comment",
  );

  expect(removed).toBe(false);
  expect(remove).not.toHaveBeenCalled();
});

it("serializes a new draft write after an in-flight deletion", async () => {
  let finishDelete!: () => void;
  const remove = jest.fn(
    () =>
      new Promise<void>((resolve) => {
        finishDelete = resolve;
      }),
  );
  jest.mocked(getDesktopBridge).mockReturnValue({
    draftWrite: write,
    draftRead: jest.fn().mockResolvedValue({ html: "Submitted" }),
    draftDelete: remove,
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);

  const removal = removeDesktopDraft(scope, "comment:queued");
  await new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
  const save = writeDesktopDraft(scope, "comment:queued", { html: "Newer" });
  expect(write).not.toHaveBeenCalled();

  finishDelete();
  await Promise.all([removal, save]);

  expect(write).toHaveBeenCalledWith(
    expect.objectContaining({ payload: { html: "Newer" } }),
  );
  await expect(flushDesktopDrafts()).resolves.toBeUndefined();
});

it("retains a failing newer write when an older submission queues conditional deletion", async () => {
  let rejectWrite!: (error: Error) => void;
  write.mockImplementationOnce(
    () =>
      new Promise<void>((_, reject) => {
        rejectWrite = reject;
      }),
  );
  const remove = jest.fn();
  jest.mocked(getDesktopBridge).mockReturnValue({
    draftWrite: write,
    draftRead: jest.fn().mockResolvedValue({ html: "Submitted" }),
    draftDelete: remove,
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);

  const save = writeDesktopDraft(scope, "comment:failed-newer", {
    html: "Newer unsaved",
  });
  const removal = removeDesktopDraft(
    scope,
    "comment:failed-newer",
    (payload) => (payload as { html: string }).html === "Submitted",
  );
  const rejected = expect(save).rejects.toThrow("Disk unavailable");
  await new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
  rejectWrite(new Error("Disk unavailable"));
  await rejected;
  await expect(removal).resolves.toBe(false);

  expect(remove).not.toHaveBeenCalled();
  await expect(flushDesktopDrafts()).rejects.toThrow("could not be saved");
  await retryDesktopDrafts(scope);
  expect(write).toHaveBeenLastCalledWith(
    expect.objectContaining({ payload: { html: "Newer unsaved" } }),
  );
  await expect(flushDesktopDrafts()).resolves.toBeUndefined();
});
