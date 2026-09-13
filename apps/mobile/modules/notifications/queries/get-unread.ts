import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getUnreadNotifications = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<number>>(
    "notifications/unread-count",
    { signal },
  );
  return response.data ?? 0;
};
