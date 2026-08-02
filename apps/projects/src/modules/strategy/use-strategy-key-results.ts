import { useMemo } from "react";
import { useQueries } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getKeyResults } from "@/modules/objectives/queries/get-key-results";
import type { KeyResult, Objective } from "@/modules/objectives/types";

export const useStrategyKeyResults = (objectives: Objective[]) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const objectivesWithKeyResults = useMemo(
    () => objectives.filter(({ keyResultCount }) => keyResultCount > 0),
    [objectives],
  );
  const queries = useQueries({
    queries: objectivesWithKeyResults.map((objective) => ({
      queryKey: objectiveKeys.keyResults(workspaceSlug, objective.id),
      queryFn: () =>
        getKeyResults(objective.id, { session: session!, workspaceSlug }),
      enabled: Boolean(session),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    })),
  });

  return useMemo(() => {
    const keyResultsByObjective = new Map<string, KeyResult[]>();
    objectivesWithKeyResults.forEach((objective, index) => {
      keyResultsByObjective.set(objective.id, queries[index]?.data ?? []);
    });

    return {
      isPending: queries.some(({ isPending }) => isPending),
      keyResultsByObjective,
    };
  }, [objectivesWithKeyResults, queries]);
};
