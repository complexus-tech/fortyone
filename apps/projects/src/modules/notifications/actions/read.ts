"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const readNotification = async (notificationId: string) => {
  try {
    const session = await auth();
    await put(`notifications/${notificationId}/read`, {}, session!);
  } catch (error) {
    return getApiError(error);
  }
};
