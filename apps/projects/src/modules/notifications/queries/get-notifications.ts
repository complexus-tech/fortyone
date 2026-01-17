import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AppNotification } from "../types";

export const getNotifications = async (ctx: WorkspaceCtx) => {
  try {
    const res = await get<ApiResponse<AppNotification[]>>(
      "notifications",
      ctx,
    );
    return res.data ?? [];
  } catch (error) {
    return [];
  }
};
