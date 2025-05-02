"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteNotification = async (notificationId: string) => {
  try {
    const session = await auth();
    await remove(`notifications/${notificationId}`, session!);
  } catch (error) {
    return getApiError(error);
  }
};
