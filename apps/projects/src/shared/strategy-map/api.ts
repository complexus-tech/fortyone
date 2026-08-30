import { get, put } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import {
  normalizeStrategyMap,
  type StrategyMapResponse,
} from "./normalize-strategy-map";

export const getStrategyMap = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<StrategyMapResponse>>(
    "strategy-map",
    ctx,
  );
  return normalizeStrategyMap(response.data);
};

export const alignObjective = (
  objectiveId: string,
  pillarId: string | null,
  ctx: WorkspaceCtx,
) =>
  put<{ pillarId: string | null }, ApiResponse<null>>(
    `strategy-map/objectives/${objectiveId}`,
    { pillarId },
    ctx,
  );
