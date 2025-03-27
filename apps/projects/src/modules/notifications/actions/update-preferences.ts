"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import type { UpdateNotificationPreferences } from "../types";

export const updateNotificationPreferences = async (
  preferences: UpdateNotificationPreferences,
  type:
    | "story_update"
    | "objective_update"
    | "comment_reply"
    | "mention"
    | "key_result_update"
    | "story_comment",
) => {
  try {
    await put(`notification-preferences/${type}`, preferences);
  } catch (error) {
    return getApiError(error);
  }
};
