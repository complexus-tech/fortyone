import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { NotificationPreferences } from "../types";

export const getNotificationPreferences = async (ctx: WorkspaceCtx) => {
  const res = await get<ApiResponse<NotificationPreferences>>(
    "notification-preferences",
    ctx,
  );
  return res.data!;
};
