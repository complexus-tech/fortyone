import { put } from "@/lib/http";
import { getApiError } from "@/utils";

export const readNotification = async (notificationId: string) => {
  try {
    await put(`notifications/${notificationId}/read`, {});
  } catch (error) {
    return getApiError(error);
  }
};
