"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteReadNotifications = async (workspaceSlug: string) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    await remove(`notifications/read`, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
