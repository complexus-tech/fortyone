import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getUnreadNotificationsStrict = async (ctx: WorkspaceCtx) => {
  const res = await get<ApiResponse<number>>("notifications/unread-count", ctx);
  if (res.error?.message) throw new Error(res.error.message);
  return res.data ?? 0;
};

export const getUnreadNotifications = async (ctx: WorkspaceCtx) => {
  try {
    return await getUnreadNotificationsStrict(ctx);
  } catch {
    return 0;
  }
};
