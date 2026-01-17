"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const markUnread = async (
  notificationId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    await put(`notifications/${notificationId}/unread`, {}, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
