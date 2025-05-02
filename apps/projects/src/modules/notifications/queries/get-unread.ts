import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getUnreadNotifications = async (session: Session) => {
  try {
    const res = await get<ApiResponse<number>>(
      "notifications/unread-count",
      session,
    );
    return res.data ?? 0;
  } catch (error) {
    return 0;
  }
};
