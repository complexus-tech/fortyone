import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getUnreadNotifications = async () => {
  const response = await get<ApiResponse<number>>("notifications/unread-count");
  return response.data ?? 0;
};
