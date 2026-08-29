import type { StrategicPillar, StrategyMap } from "./types";

export type StrategyMapResponse = Omit<StrategyMap, "pillars"> & {
  pillars?: Array<
    Omit<StrategicPillar, "objectiveIds"> & {
      objectiveIds?: string[] | null;
    }
  > | null;
};

export const normalizeStrategyMap = (
  strategy?: StrategyMapResponse | null,
): StrategyMap => ({
  ultimateGoal: strategy?.ultimateGoal ?? "",
  description: strategy?.description ?? null,
  pillars: (strategy?.pillars ?? []).map((pillar) => ({
    ...pillar,
    objectiveIds: pillar.objectiveIds ?? [],
  })),
});
