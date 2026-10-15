import { randomUUID } from "node:crypto";
import { waitFor } from "@testing-library/react";
import { getDesktopBridge } from "./bridge";
import {
  DESKTOP_ATTACHMENT_LIMIT,
  flushDesktopAttachments,
  stageDesktopAttachment,
} from "./attachments";
import { DESKTOP_OUTBOX_CHANGED } from "./outbox-types";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./session-binding", () => ({
  bindDesktopSession: jest.fn().mockResolvedValue(true),
}));
const scope = { accountId: "alice", workspaceId: "one", workspaceSlug: "acme" };
const stage = jest.fn();
beforeEach(() => {
  jest.clearAllMocks();
  Object.defineProperty(crypto, "randomUUID", {
    configurable: true,
    value: randomUUID,
  });
  stage.mockImplementation(async ({ id, fileName }) => ({
    operationId: id,
    attachment: { id, filename: fileName, pending: true, url: "" },
  }));
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue({ stageAttachment: stage } as unknown as NonNullable<
      ReturnType<typeof getDesktopBridge>
    >);
});

it("reuses the receipt identity after a failed response and scopes a file separately per workspace", async () => {
  const file = new File(["private file"], "notes.txt", { type: "text/plain" });
  stage.mockRejectedValueOnce(new Error("Response lost"));
  await expect(stageDesktopAttachment(scope, "story", file)).rejects.toThrow(
    "Response lost",
  );
  const id = stage.mock.calls[0][0].id;
  await stageDesktopAttachment(scope, "story", file);
  expect(stage).toHaveBeenLastCalledWith(
    expect.objectContaining({
      id,
      data: btoa("private file"),
      accountId: "alice",
      workspaceId: "one",
    }),
  );
  await stageDesktopAttachment({ ...scope, workspaceId: "two" }, "story", file);
  expect(stage.mock.calls[2][0].id).not.toBe(id);
});

it("rejects oversized files before reading or crossing the bridge", async () => {
  const file = new File(["small"], "large.bin");
  Object.defineProperty(file, "size", { value: DESKTOP_ATTACHMENT_LIMIT + 1 });
  await expect(stageDesktopAttachment(scope, "story", file)).rejects.toThrow(
    "10 MB",
  );
  expect(stage).not.toHaveBeenCalled();
});

it("holds the durability gate and only publishes an outbox change after native storage succeeds", async () => {
  let release!: (value: unknown) => void;
  stage.mockImplementation(
    () =>
      new Promise((resolve) => {
        release = resolve;
      }),
  );
  const changed = jest.fn();
  window.addEventListener(DESKTOP_OUTBOX_CHANGED, changed);
  const save = stageDesktopAttachment(
    scope,
    "story",
    new File(["hello"], "hello.txt"),
  );
  let flushed = false;
  const flush = flushDesktopAttachments().then(() => {
    flushed = true;
  });
  await waitFor(() => {
    expect(stage).toHaveBeenCalledTimes(1);
  });
  expect(flushed).toBe(false);
  expect(changed).not.toHaveBeenCalled();
  release({ operationId: "op", attachment: { id: "file", pending: true } });
  await Promise.all([save, flush]);
  expect(flushed).toBe(true);
  expect(changed).toHaveBeenCalledTimes(1);
  window.removeEventListener(DESKTOP_OUTBOX_CHANGED, changed);
});

it("does not silently fall back to a web upload when the desktop capability is missing", async () => {
  jest.mocked(getDesktopBridge).mockReturnValue(null);
  await expect(
    stageDesktopAttachment(scope, "story", new File(["x"], "file.txt")),
  ).rejects.toThrow("Update the desktop app");
  expect(stage).not.toHaveBeenCalled();
});
