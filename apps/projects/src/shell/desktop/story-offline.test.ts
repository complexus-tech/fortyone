import { randomUUID } from "node:crypto";
import { QueryClient } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import { getDesktopBridge } from "./bridge";
import {
  commentDesktopStory,
  createDesktopStory,
  localStory,
  updateDesktopStory,
} from "./story-offline";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./session-binding", () => ({
  bindDesktopSession: jest.fn().mockResolvedValue(true),
}));
const scope = {
  accountId: "alice",
  workspaceId: "workspace",
  workspaceSlug: "acme",
};
const version = "2026-10-15T10:00:00.000Z";
const enqueue = jest.fn();
beforeEach(() => {
  jest.clearAllMocks();
  Object.defineProperty(crypto, "randomUUID", {
    configurable: true,
    value: randomUUID,
  });
  enqueue.mockImplementation(async (input) => ({
    operationId: input.id,
    resourceId: input.resourceId,
    pending: true,
  }));
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue({ enqueueOperation: enqueue } as unknown as NonNullable<
      ReturnType<typeof getDesktopBridge>
    >);
});

it("uses a server version for pending updates and pending-create only for provisional stories", async () => {
  const client = new QueryClient();
  const key = storyKeys.detail("acme", "item");
  client.setQueryData(key, {
    ...localStory({}, "item", scope, version),
    desktopCreated: false,
    desktopPending: true,
  });
  await updateDesktopStory(scope, client, "item", {
    id: "routing-only",
    title: "Edit",
  });
  expect(enqueue).toHaveBeenLastCalledWith(
    expect.objectContaining({
      resourceId: "item",
      baseVersion: version,
      payload: { title: "Edit" },
    }),
  );
  client.setQueryData(key, localStory({}, "item", scope, version));
  await updateDesktopStory(scope, client, "item", { title: "Edit" });
  expect(enqueue).toHaveBeenLastCalledWith(
    expect.objectContaining({ baseVersion: "pending-create:item" }),
  );
});

it("keeps recovered draft version metadata out of the mutable resource payload", async () => {
  const client = new QueryClient();
  client.setQueryData(
    storyKeys.detail("acme", "item"),
    localStory({}, "item", scope, version),
  );
  await updateDesktopStory(
    scope,
    client,
    "item",
    { title: "Recovered" },
    "older-original-version",
  );
  expect(enqueue).toHaveBeenCalledWith(
    expect.objectContaining({
      baseVersion: "older-original-version",
      payload: { title: "Recovered" },
    }),
  );
});

it("atomically consumes the designated new-story draft and respects already acknowledged retries", async () => {
  enqueue.mockResolvedValue({
    operationId: "operation",
    resourceId: "server",
    pending: false,
    resource: {
      title: "Created",
      sequenceId: 7,
      teamCode: "ENG",
      updatedAt: version,
    },
  });
  const result = await createDesktopStory(
    scope,
    { idempotencyKey: "operation", title: "Created" },
    "new-story:team",
  );
  expect(enqueue).toHaveBeenCalledWith(
    expect.objectContaining({
      id: "operation",
      consumedDraftKey: "new-story:team",
      accountId: "alice",
      workspaceId: "workspace",
    }),
  );
  expect(result).toMatchObject({
    id: "server",
    sequenceId: 7,
    desktopCreated: false,
    desktopPending: false,
  });
});

it("reuses operation IDs after an ambiguous IPC failure without marking acknowledged comments pending", async () => {
  const payload = { title: "Retried create" };
  enqueue.mockRejectedValueOnce(new Error("IPC response lost"));
  await expect(createDesktopStory(scope, payload)).rejects.toThrow();
  await createDesktopStory(scope, payload);
  expect(enqueue.mock.calls[0][0].id).toBe(enqueue.mock.calls[1][0].id);
  enqueue.mockResolvedValue({ operationId: "comment-op", resourceId: "server-comment", pending: false, resource: { comment: "Saved on server" } });
  const result = await commentDesktopStory(scope, null, "item", { comment: "Saved on server", mentions: [] });
  expect(result).toMatchObject({ desktopPending: false, data: { id: "server-comment", comment: "Saved on server" } });
});
