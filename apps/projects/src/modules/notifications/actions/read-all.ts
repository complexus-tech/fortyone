"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";

export const readAllNotifications = async () => {
  try {
    await put(`notifications/read-all`, {});
  } catch (error) {
    return getApiError(error);
  }
};
