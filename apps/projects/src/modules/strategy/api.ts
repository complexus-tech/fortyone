import { post, put, remove } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  NewStrategicPillar,
  StrategicPillar,
  StrategyUpdate,
  UpdateStrategicPillar,
} from "./types";

export { alignObjective, getStrategyMap } from "@/shared/strategy-map/api";

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

export const deleteStrategicPillar = (pillarId: string, ctx: WorkspaceCtx) =>
  remove<ApiResponse<null>>(`strategy-map/pillars/${pillarId}`, ctx);
