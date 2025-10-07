import { put } from "@/lib/http";

export const markUnread = async (notificationId: string) => {
  return put(`notifications/${notificationId}/unread`, {});
};
