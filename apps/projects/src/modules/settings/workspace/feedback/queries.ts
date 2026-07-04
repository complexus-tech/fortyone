import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { FeedbackPortal } from "./types";

export const getFeedbackPortals = async (
  ctx: WorkspaceCtx,
): Promise<FeedbackPortal[]> => {
  const response = await get<ApiResponse<FeedbackPortal[]>>(
    "feedback/portals",
    ctx,
  );
  return response.data ?? [];
};
