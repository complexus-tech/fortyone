import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AppNotification } from "../types";

export const getNotifications = async () => {
  const response = await get<ApiResponse<AppNotification[]>>("notifications");
  return response.data ?? [];
};
