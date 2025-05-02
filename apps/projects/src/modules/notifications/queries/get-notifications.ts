import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AppNotification } from "../types";

export const getNotifications = async (session: Session) => {
  try {
    const res = await get<ApiResponse<AppNotification[]>>(
      "notifications",
      session,
    );
    return res.data ?? [];
  } catch (error) {
    return [];
  }
};
