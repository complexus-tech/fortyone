import { remove } from "@/lib/http";

export const deleteReadNotifications = async () => {
  return remove(`notifications/read`);
};
