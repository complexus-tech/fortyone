import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ObjectiveStatus } from "../types";

export const getObjectiveStatuses = async (ctx: WorkspaceCtx) => {
  const statuses = await get<ApiResponse<ObjectiveStatus[]>>(
    "objective-statuses",
    ctx,
  );
  return statuses.data!;
};
