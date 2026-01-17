import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, WorkspaceSettings } from "@/types";

export const getWorkspaceSettings = async (ctx: WorkspaceCtx) => {
  const settings = await get<ApiResponse<WorkspaceSettings>>("settings", ctx);
  return settings.data!;
};
