import { put } from "@/lib/http";

export const readAllNotifications = async () => {
  await put("notifications/read-all", {});
};
