"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { UpdateNotificationPreferences } from "../types";

export const updateNotificationPreferences = async (
  preferences: UpdateNotificationPreferences,
  type:
    | "story_update"
    | "objective_update"
    | "comment_reply"
    | "mention"
    | "key_result_update"
    | "story_comment"
    | "reminders",
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    await put(`notification-preferences/${type}`, preferences, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
