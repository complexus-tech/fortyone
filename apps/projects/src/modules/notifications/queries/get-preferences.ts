import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { NotificationPreferences } from "../types";

export const getNotificationPreferences = async () => {
  const res = await get<ApiResponse<NotificationPreferences>>(
    "notification-preferences",
  );
  return res.data!;
};
