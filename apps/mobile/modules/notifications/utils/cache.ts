import type { InfiniteData } from "@tanstack/react-query";
import type { NotificationsPage } from "../types";

export type NotificationsData = InfiniteData<NotificationsPage, number>;
export type NotificationChange =
  | { type: "read" | "unread" | "delete"; id: string }
  | { type: "read-all" | "delete-all" | "delete-read" };

export const updateNotificationsCache = (
  data: NotificationsData | undefined,
  unreadCount: number | undefined,
  change: NotificationChange,
  now = new Date().toISOString(),
) => {
  let unreadDelta = 0;
  const counted = new Set<string>();
  const pages = data?.pages.map((page) => ({
    ...page,
    notifications: page.notifications.flatMap((notification) => {
      if ("id" in change && notification.id !== change.id)
        return [notification];
      const unread = notification.readAt === null;
      if (!counted.has(notification.id)) {
        if ((change.type === "read" || change.type === "delete") && unread)
          unreadDelta--;
        if (change.type === "unread" && !unread) unreadDelta++;
        counted.add(notification.id);
      }
      if (
        change.type === "delete" ||
        change.type === "delete-all" ||
        (change.type === "delete-read" && !unread)
      )
        return [];
      if (change.type === "read" || change.type === "read-all") {
        return [{ ...notification, readAt: notification.readAt ?? now }];
      }
      if (change.type === "unread") return [{ ...notification, readAt: null }];
      return [notification];
    }),
  }));
  return {
    data: data && pages ? { ...data, pages } : data,
    unreadCount:
      change.type === "read-all" || change.type === "delete-all"
        ? 0
        : unreadCount === undefined
          ? undefined
          : Math.max(0, unreadCount + unreadDelta),
  };
};
