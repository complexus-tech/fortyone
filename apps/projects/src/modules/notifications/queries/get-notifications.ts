"use server";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AppNotification } from "../types";

export const getNotifications = async () => {
  try {
    const res = await get<ApiResponse<AppNotification[]>>("notifications");
    return res.data ?? [];
  } catch (error) {
    return [];
  }
};
