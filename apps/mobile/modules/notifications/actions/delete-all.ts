import { remove } from "@/lib/http";

export const deleteAllNotifications = async () => {
  return remove(`notifications`);
};
