import { useCallback, useMemo, useState } from "react";
import { useQueries, type UseQueryResult } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getKeyResults } from "@/modules/objectives/queries/get-key-results";
import type { KeyResult, Objective } from "@/modules/objectives/types";

export const getStrategyKeyResultQueryObjectives = (
  objectives: readonly Objective[],
  activeObjectiveIds: ReadonlySet<string>,
  retainedObjectiveIds?: ReadonlySet<string>,
) =>
  objectives.filter(
    ({ id, keyResultCount }) =>
      keyResultCount > 0 &&
      (activeObjectiveIds.has(id) || retainedObjectiveIds?.has(id)),
  );

export const useStrategyKeyResults = (
  objectives: Objective[],
  activeObjectiveIds: ReadonlySet<string>,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const [retainedQueries, setRetainedQueries] = useState(() => ({
    objectiveIds: activeObjectiveIds,
    workspaceSlug,
  }));
  const objectivesWithKeyResults = useMemo(
    () =>
      getStrategyKeyResultQueryObjectives(
        objectives,
        activeObjectiveIds,
        retainedQueries.workspaceSlug === workspaceSlug
          ? retainedQueries.objectiveIds
          : undefined,
      ),
    [activeObjectiveIds, objectives, retainedQueries, workspaceSlug],
  );
  const retainedObjectiveIds = useMemo(
    () => new Set(objectivesWithKeyResults.map(({ id }) => id)),
    [objectivesWithKeyResults],
  );

  // Retain the graph data after it leaves the render window. Removing its
  // query also removed its key-result nodes and connectors, even while part of
  // the connector could still cross the viewport. Prune deleted objectives and
  // reset on workspace changes before rendering the next graph.
  if (
    retainedQueries.workspaceSlug !== workspaceSlug ||
    retainedQueries.objectiveIds.size !== retainedObjectiveIds.size ||
    objectivesWithKeyResults.some(
      ({ id }) => !retainedQueries.objectiveIds.has(id),
    )
  ) {
    setRetainedQueries({ objectiveIds: retainedObjectiveIds, workspaceSlug });
  }

  const combine = useCallback(
    (queries: UseQueryResult<KeyResult[]>[]) => ({
      entries: objectivesWithKeyResults.map(({ id }, index) => ({
        objectiveId: id,
        keyResults: queries[index]?.data,
      })),
      isPending: queries.some(
        ({ isPending }, index) =>
          isPending &&
          activeObjectiveIds.has(objectivesWithKeyResults[index].id),
      ),
      workspaceSlug,
    }),
    [activeObjectiveIds, objectivesWithKeyResults, workspaceSlug],
  );
  const { entries, isPending } = useQueries({
    queries: objectivesWithKeyResults.map((objective) => ({
      queryKey: objectiveKeys.keyResults(workspaceSlug, objective.id),
      queryFn: () =>
        getKeyResults(objective.id, { session: session!, workspaceSlug }),
      enabled: Boolean(session) && activeObjectiveIds.has(objective.id),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    })),
    combine,
  });

  // useQueries returns a new result array on each render. Its combined plain
  // snapshot is structurally shared, so camera updates cannot invalidate layout
  // maps when the visible query window and key-result data have not changed.
  const maps = useMemo(() => {
    const keyResultsByObjective = new Map<string, KeyResult[]>();
    const loadedObjectiveIds = new Set<string>();
    entries.forEach(({ objectiveId, keyResults }) => {
      if (keyResults !== undefined) loadedObjectiveIds.add(objectiveId);
      keyResultsByObjective.set(objectiveId, keyResults ?? []);
    });

    return {
      keyResultsByObjective,
      loadedObjectiveIds,
    };
  }, [entries]);

  return { ...maps, isPending };
};
