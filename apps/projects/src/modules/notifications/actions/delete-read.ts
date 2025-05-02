"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteReadNotifications = async () => {
  try {
    const session = await auth();
    await remove(`notifications/read`, session!);
  } catch (error) {
    return getApiError(error);
  }
};
