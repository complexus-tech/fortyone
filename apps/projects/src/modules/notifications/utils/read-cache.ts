import type { InfiniteData } from "@tanstack/react-query";
import type { ReadNotificationInput } from "lib/src/notification-read";
import type { AppNotification, NotificationsPage } from "../types";

export type NotificationReadCache =
  | AppNotification[]
  | InfiniteData<NotificationsPage, number>;

export const isNotificationReadCache = (
  value: unknown,
): value is NotificationReadCache => {
  if (Array.isArray(value)) return true;
  if (!value || typeof value !== "object" || !("pages" in value)) return false;
  return (
    Array.isArray(value.pages) &&
    value.pages.every(
      (page: unknown) =>
        page !== null &&
        typeof page === "object" &&
        "notifications" in page &&
        Array.isArray(page.notifications),
    )
  );
};

export const markNotificationReadInCache = (
  data: NotificationReadCache,
  { id, observedCreatedAt }: ReadNotificationInput,
  readAt: string,
) => {
  const update = (notifications: AppNotification[]) => {
    const next = notifications.map((notification) => {
      if (
        notification.id !== id ||
        notification.readAt !== null ||
        (observedCreatedAt && notification.createdAt !== observedCreatedAt)
      ) {
        return notification;
      }
      return { ...notification, readAt };
    });
    return next.some(
      (notification, index) => notification !== notifications[index],
    )
      ? next
      : notifications;
  };
  if (Array.isArray(data)) {
    const next = update(data);
    return { data: next, changed: next !== data };
  }
  const pages = data.pages.map((page) => {
    const notifications = update(page.notifications);
    return notifications === page.notifications
      ? page
      : { ...page, notifications };
  });
  const changed = pages.some((page, index) => page !== data.pages[index]);
  return { data: changed ? { ...data, pages } : data, changed };
};
