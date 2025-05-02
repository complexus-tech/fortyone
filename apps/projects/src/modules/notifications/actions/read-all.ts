"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const readAllNotifications = async () => {
  try {
    const session = await auth();
    await put(`notifications/read-all`, {}, session!);
  } catch (error) {
    return getApiError(error);
  }
};
