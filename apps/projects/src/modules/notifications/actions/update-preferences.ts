import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  NotificationType,
  UpdateNotificationPreferences,
} from "../types";

export const updateNotificationPreferences = async (
  preferences: UpdateNotificationPreferences,
  type: NotificationType,
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
