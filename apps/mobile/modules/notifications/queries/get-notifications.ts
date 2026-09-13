import { parseNotificationsPage } from "../utils/response";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { NotificationsPage } from "../types";

export const getNotifications = async (page = 1, signal?: AbortSignal) => {
  const response = await get<ApiResponse<NotificationsPage>>(
    `notifications?page=${page}&pageSize=25`,
    { signal },
  );
  if (response.error?.message) throw new Error(response.error.message);
  return parseNotificationsPage(response.data);
};
