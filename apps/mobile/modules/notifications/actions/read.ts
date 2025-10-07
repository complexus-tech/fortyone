import { put } from "@/lib/http";

export const readNotification = async (notificationId: string) => {
  return put(`notifications/${notificationId}/read`, {});
};
