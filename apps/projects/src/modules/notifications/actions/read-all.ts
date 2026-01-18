"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const readAllNotifications = async (workspaceSlug: string) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    await put(`notifications/read-all`, {}, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
