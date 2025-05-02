"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const markUnread = async (notificationId: string) => {
  try {
    const session = await auth();
    await put(`notifications/${notificationId}/unread`, {}, session!);
  } catch (error) {
    return getApiError(error);
  }
};
