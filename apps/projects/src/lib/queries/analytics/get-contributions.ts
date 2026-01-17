import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Contribution } from "@/types";

export const getContributions = async (ctx: WorkspaceCtx) => {
  const contributions = await get<ApiResponse<Contribution[]>>(
    "analytics/contributions",
    ctx,
  );
  return contributions.data!;
};
