import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { NotificationPreferences } from "../types";

export const getNotificationPreferences = async () => {
  const response = await get<ApiResponse<NotificationPreferences>>(
    "notification-preferences",
  );
  if (response.error?.message) throw new Error(response.error.message);
  if (!response.data)
    throw new Error("Notification preferences are unavailable.");
  return response.data;
};
