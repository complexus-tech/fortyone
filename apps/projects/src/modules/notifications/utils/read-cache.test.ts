/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getNotificationReadPath } from "lib/src/notification-read";
import type { AppNotification, NotificationsPage } from "../types";
import {
  isNotificationReadCache,
  markNotificationReadInCache,
} from "./read-cache";

const createdAt = "2026-10-03T20:00:00.123456789Z";
const notification: AppNotification = {
  id: "notification-1",
  recipientId: "recipient",
  workspaceId: "workspace",
  type: "story_update",
  entityType: "story",
  entityId: "story",
  actorId: "actor",
  title: "Updated",
  message: { template: "Updated", variables: {} },
  createdAt,
  readAt: null,
};
const page: NotificationsPage = {
  notifications: [notification],
  pagination: { page: 1, pageSize: 25, hasMore: false, nextPage: 2 },
};

describe("notification read versions", () => {
  it("URL-encodes the original timestamp and preserves older callers", () => {
    const observedCreatedAt = "2026-10-03T20:00:00.123456789+02:00";
    const path = getNotificationReadPath("notification/1", observedCreatedAt);
    expect(path.split("?")[0]).toBe("notifications/notification%2F1/read");
    expect(
      new URL(path, "https://fortyone.app/").searchParams.get(
        "observedCreatedAt",
      ),
    ).toBe(observedCreatedAt);
    expect(getNotificationReadPath("notification-1")).toBe(
      "notifications/notification-1/read",
    );
  });

  it("does not consume a newer unread row from either inbox cache shape", () => {
    const input = {
      id: notification.id,
      observedCreatedAt: "2026-10-03T19:00:00Z",
    };
    const flat = [notification];
    const infinite = { pages: [page], pageParams: [1] };
    expect(markNotificationReadInCache(flat, input, "now")).toEqual({
      data: flat,
      changed: false,
    });
    expect(markNotificationReadInCache(infinite, input, "now").data).toBe(
      infinite,
    );
    expect(notification.readAt).toBeNull();
  });

  it("marks matching copies on every page once and retains pagination", () => {
    const data = {
      pages: [page, { ...page, notifications: [notification] }],
      pageParams: [1, 2],
    };
    const input = { id: notification.id, observedCreatedAt: createdAt };
    const next = markNotificationReadInCache(data, input, "now");
    expect(next.changed).toBe(true);
    expect(
      "pages" in next.data &&
        next.data.pages.map((item) => item.notifications[0].readAt),
    ).toEqual(["now", "now"]);
    expect("pageParams" in next.data && next.data.pageParams).toEqual([1, 2]);
    expect(markNotificationReadInCache(next.data, input, "later")).toEqual({
      data: next.data,
      changed: false,
    });
  });

  it("ignores unread-count and preference caches", () => {
    expect(isNotificationReadCache(3)).toBe(false);
    expect(isNotificationReadCache({ preferences: {} })).toBe(false);
    expect(isNotificationReadCache({ pages: [{ records: [] }] })).toBe(false);
  });
});
