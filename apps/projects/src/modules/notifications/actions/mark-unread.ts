"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";

export const markUnread = async (notificationId: string) => {
  try {
    await put(`notifications/${notificationId}/unread`, {});
  } catch (error) {
    return getApiError(error);
  }
};
