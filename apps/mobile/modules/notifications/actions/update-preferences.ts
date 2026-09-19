import { put } from "@/lib/http";
import type { NotificationType, UpdateNotificationPreferences } from "../types";

export const updateNotificationPreferences = (
  type: NotificationType,
  preferences: UpdateNotificationPreferences,
) =>
  put<UpdateNotificationPreferences, void>(
    `notification-preferences/${type}`,
    preferences,
  );
