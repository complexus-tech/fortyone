"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteReadNotifications = async () => {
  try {
    await remove(`notifications/read`);
  } catch (error) {
    return getApiError(error);
  }
};
