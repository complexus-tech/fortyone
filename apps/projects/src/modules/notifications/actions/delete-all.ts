import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteAllNotifications = async () => {
  try {
    await remove(`notifications`);
  } catch (error) {
    return getApiError(error);
  }
};
