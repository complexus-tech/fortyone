"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteNotification = async (notificationId: string) => {
  try {
    await remove(`notifications/${notificationId}`);
  } catch (error) {
    return getApiError(error);
  }
};
