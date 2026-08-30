import type { Objective } from "@/shared/objectives/types";
import {
  getKeyResultProgress,
  type KeyResultProgressInput,
} from "@/shared/key-results/progress";

export const getObjectiveProgress = (
  objective: Objective,
  keyResults: readonly KeyResultProgressInput[] = [],
) => {
  if (objective.keyResultCount > 0 || keyResults.length > 0) {
    if (keyResults.length === 0) return 0;

    return Math.round(
      keyResults.reduce(
        (total, keyResult) => total + getKeyResultProgress(keyResult),
        0,
      ) / keyResults.length,
    );
  }

  const total = objective.stats?.total ?? 0;
  const completed = objective.stats?.completed ?? 0;
  return total > 0 ? Math.round((completed / total) * 100) : 0;
};

export const getCompleteStrategyMapAverageProgress = (
  objectives: readonly Objective[],
  keyResultsByObjective: ReadonlyMap<string, readonly KeyResultProgressInput[]>,
  loadedObjectiveIds: ReadonlySet<string>,
): number | null => {
  if (objectives.length === 0) return 0;

  let totalProgress = 0;
  for (const objective of objectives) {
    if (objective.keyResultCount > 0 && !loadedObjectiveIds.has(objective.id)) {
      return null;
    }

    totalProgress += getObjectiveProgress(
      objective,
      keyResultsByObjective.get(objective.id),
    );
  }

  return Math.round(totalProgress / objectives.length);
};
