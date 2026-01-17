import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getUnreadNotifications = async (ctx: WorkspaceCtx) => {
  try {
    const res = await get<ApiResponse<number>>("notifications/unread-count", ctx);
    return res.data ?? 0;
  } catch (error) {
    return 0;
  }
};
