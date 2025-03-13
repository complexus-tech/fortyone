"use server";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getUnreadNotifications = async () => {
  try {
    const res = await get<ApiResponse<number>>("notifications/unread-count");
    return res.data ?? 0;
  } catch (error) {
    return 0;
  }
};
