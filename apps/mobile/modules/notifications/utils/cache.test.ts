import assert from "node:assert/strict";
import test from "node:test";
import type { AppNotification } from "../types";
import type { NotificationsData } from "./cache";
import { updateNotificationsCache } from "./cache.ts";

const item = (id: string, readAt: string | null): AppNotification => ({
  id,
  readAt,
  title: id,
  recipientId: "a",
  workspaceId: "w",
  type: "story_update",
  entityType: "story",
  entityId: "task",
  actorId: "actor",
  createdAt: "2026-09-12",
  message: {
    template: "Updated",
    variables: {
      actor: { value: "A" },
      field: { value: "title" },
      value: { value: "New" },
    },
  },
});
const data: NotificationsData = {
  pages: [
    {
      notifications: [item("one", null)],
      pagination: { page: 1, pageSize: 1, hasMore: true, nextPage: 2 },
    },
    {
      notifications: [item("two", "earlier")],
      pagination: { page: 2, pageSize: 1, hasMore: false, nextPage: 3 },
    },
  ],
  pageParams: [1, 2],
};

test("read/unread changes are idempotent and operate on every inbox page", () => {
  const first = updateNotificationsCache(
    data,
    1,
    { type: "read", id: "one" },
    "now",
  );
  assert.equal(first.unreadCount, 0);
  assert.equal(first.data?.pages[0].notifications[0].readAt, "now");
  assert.equal(
    updateNotificationsCache(first.data, first.unreadCount, {
      type: "read",
      id: "one",
    }).unreadCount,
    0,
  );
  const unread = updateNotificationsCache(data, 1, {
    type: "unread",
    id: "two",
  });
  assert.equal(unread.unreadCount, 2);
  assert.equal(
    updateNotificationsCache(unread.data, 2, { type: "unread", id: "two" })
      .unreadCount,
    2,
  );
  assert.deepEqual(unread.data?.pageParams, [1, 2]);
});

test("delete and read-all preserve pagination and count unloaded notifications correctly", () => {
  const deleted = updateNotificationsCache(data, 8, {
    type: "delete",
    id: "one",
  });
  assert.equal(deleted.unreadCount, 7);
  assert.deepEqual(deleted.data?.pages[0].notifications, []);
  assert.deepEqual(deleted.data?.pages[0].pagination, data.pages[0].pagination);
  const all = updateNotificationsCache(data, 8, { type: "read-all" }, "now");
  assert.equal(all.unreadCount, 0);
  assert.equal(all.data?.pages[1].notifications[0].readAt, "earlier");
});

test("deleting a duplicate notification removes every copy but decrements unread only once", () => {
  const repeated = {
    ...data,
    pages: [
      data.pages[0],
      {
        ...data.pages[1],
        notifications: [item("one", null), item("two", "earlier")],
      },
    ],
  };
  const deleted = updateNotificationsCache(repeated, 8, {
    type: "delete",
    id: "one",
  });
  assert.equal(deleted.unreadCount, 7);
  assert.deepEqual(
    deleted.data?.pages.flatMap((page) =>
      page.notifications.map((notification) => notification.id),
    ),
    ["two"],
  );
  assert.deepEqual(deleted.data?.pageParams, data.pageParams);
  assert.equal(
    updateNotificationsCache(deleted.data, 7, { type: "delete", id: "one" })
      .unreadCount,
    7,
  );
  assert.equal(
    updateNotificationsCache(deleted.data, 7, { type: "delete", id: "two" })
      .unreadCount,
    7,
  );
});
