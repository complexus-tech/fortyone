import type { StrategyMap } from "./types";

export const alignObjectiveInStrategy = (
  strategy: StrategyMap,
  objectiveId: string,
  pillarId: string | null,
): StrategyMap => ({
  ...strategy,
  pillars: strategy.pillars.map((pillar) => {
    const objectiveIds = pillar.objectiveIds.filter((id) => id !== objectiveId);
    if (pillar.id !== pillarId) return { ...pillar, objectiveIds };

    return {
      ...pillar,
      objectiveIds: [...objectiveIds, objectiveId],
    };
  }),
});
