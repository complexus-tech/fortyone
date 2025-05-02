import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { NotificationPreferences } from "../types";

export const getNotificationPreferences = async (session: Session) => {
  const res = await get<ApiResponse<NotificationPreferences>>(
    "notification-preferences",
    session,
  );
  return res.data!;
};
