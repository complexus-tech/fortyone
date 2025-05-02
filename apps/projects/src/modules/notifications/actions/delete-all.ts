"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteAllNotifications = async () => {
  try {
    const session = await auth();
    await remove(`notifications`, session!);
  } catch (error) {
    return getApiError(error);
  }
};
