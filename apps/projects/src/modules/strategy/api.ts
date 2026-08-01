import { get, post, put, remove } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  NewStrategicPillar,
  StrategicPillar,
  StrategyMap,
  StrategyUpdate,
  UpdateStrategicPillar,
} from "./types";

export const getStrategyMap = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<StrategyMap>>("strategy-map", ctx);
  return response.data ?? { ultimateGoal: "", description: null, pillars: [] };
};

export const updateStrategy = (strategy: StrategyUpdate, ctx: WorkspaceCtx) =>
  put<StrategyUpdate, ApiResponse<null>>("strategy-map", strategy, ctx);

export const createStrategicPillar = (
  pillar: NewStrategicPillar,
  ctx: WorkspaceCtx,
) =>
  post<NewStrategicPillar, ApiResponse<StrategicPillar>>(
    "strategy-map/pillars",
    pillar,
    ctx,
  );

export const updateStrategicPillar = (
  pillarId: string,
  pillar: UpdateStrategicPillar,
  ctx: WorkspaceCtx,
) =>
  put<UpdateStrategicPillar, ApiResponse<StrategicPillar>>(
    `strategy-map/pillars/${pillarId}`,
    pillar,
    ctx,
  );

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

export const deleteStrategicPillar = (pillarId: string, ctx: WorkspaceCtx) =>
  remove<ApiResponse<null>>(`strategy-map/pillars/${pillarId}`, ctx);
