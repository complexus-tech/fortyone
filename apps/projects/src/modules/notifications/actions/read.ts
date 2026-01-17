"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const readNotification = async (
  notificationId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    await put(`notifications/${notificationId}/read`, {}, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
