import { put } from "@/lib/http";
import { getNotificationReadPath } from "lib/src/notification-read";

export const readNotification = async (
  notificationId: string,
  observedCreatedAt?: string,
) => {
  return put(getNotificationReadPath(notificationId, observedCreatedAt), {});
};
