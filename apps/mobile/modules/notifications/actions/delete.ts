import { remove } from "@/lib/http";

export const deleteNotification = async (notificationId: string) => {
  return remove(`notifications/${notificationId}`);
};
