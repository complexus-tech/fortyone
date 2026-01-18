import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Invitation } from "../types";

export const getPendingInvitations = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<Invitation[]>>("invitations", ctx);
  return response.data ?? [];
};
