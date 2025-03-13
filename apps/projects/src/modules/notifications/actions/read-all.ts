"use server";

import { post } from "@/lib/http";
import { getApiError } from "@/utils";

export const readAllNotifications = async () => {
  try {
    await post(`notifications/read-all`, {});
  } catch (error) {
    return getApiError(error);
  }
};
