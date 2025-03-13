"use server";

import { post } from "@/lib/http";
import { getApiError } from "@/utils";

export const readNotification = async (notificationId: string) => {
  try {
    await post(`notifications/${notificationId}/read`, {});
  } catch (error) {
    return getApiError(error);
  }
};
