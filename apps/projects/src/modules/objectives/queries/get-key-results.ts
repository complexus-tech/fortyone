import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { KeyResult } from "../types";

export const getKeyResults = async (objectiveId: string, ctx: WorkspaceCtx) => {
  const keyResults = await get<ApiResponse<KeyResult[]>>(
    `objectives/${objectiveId}/key-results`,
    ctx,
  );
  return keyResults.data ?? [];
};
