import type {
  InfiniteData,
  Query,
  QueryClient,
  QueryKey,
} from "@tanstack/react-query";
import type { Objective, ObjectivesPage } from "../types";
import { objectiveKeys } from "../constants";

type ObjectiveListData = Objective[] | InfiniteData<ObjectivesPage>;

export type ObjectiveCreationCacheContext = {
  optimisticId: string;
  queryKeys: QueryKey[];
};

const isPaginatedObjectiveList = (
  data: unknown,
): data is InfiniteData<ObjectivesPage> =>
  typeof data === "object" &&
  data !== null &&
  "pages" in data &&
  Array.isArray(data.pages) &&
  "pageParams" in data &&
  Array.isArray(data.pageParams) &&
  data.pages.every(
    (page: unknown) =>
      typeof page === "object" &&
      page !== null &&
      "objectives" in page &&
      Array.isArray(page.objectives),
  );

const isUnfilteredListForTeam = (queryKey: QueryKey, teamId: string) => {
  // The legacy "list" prefix also contains detail and activity keys. Only
  // known collection keys are eligible; search results reconcile on settlement.
  if (queryKey.length === 3) return true;
  if (queryKey.length === 4) return queryKey[3] === "";
  if (queryKey[3] !== teamId) return false;
  if (queryKey.length === 5) return queryKey[4] === "";
  return (
    queryKey.length === 7 && queryKey[4] === "infinite" && queryKey[5] === ""
  );
};

const appendObjective = (data: unknown, objective: Objective) => {
  if (Array.isArray(data)) return [...data, objective];
  if (!isPaginatedObjectiveList(data) || data.pages.length === 0) return data;

  return {
    ...data,
    pages: data.pages.map((page, index) =>
      index === 0
        ? { ...page, objectives: [...page.objectives, objective] }
        : page,
    ),
  };
};

export const optimisticallyCreateObjective = async (
  queryClient: QueryClient,
  workspaceSlug: string,
  objective: Objective,
): Promise<ObjectiveCreationCacheContext> => {
  const queries = queryClient.getQueryCache().findAll({
    queryKey: objectiveKeys.list(workspaceSlug),
    type: "active",
    predicate: (query: Query) =>
      isUnfilteredListForTeam(query.queryKey, objective.teamId) &&
      (Array.isArray(query.state.data) ||
        isPaginatedObjectiveList(query.state.data)),
  });
  const queryKeys = queries.map((query) => query.queryKey);

  await Promise.all(
    queryKeys.map((queryKey) =>
      queryClient.cancelQueries({ queryKey, exact: true }),
    ),
  );

  for (const queryKey of queryKeys) {
    queryClient.setQueryData(queryKey, (data: unknown) =>
      appendObjective(data, objective),
    );
  }

  return { optimisticId: objective.id, queryKeys };
};

export const settleOptimisticObjective = (
  queryClient: QueryClient,
  context: ObjectiveCreationCacheContext | undefined,
  createdObjective?: Objective,
) => {
  if (!context) return;

  const settleList = (objectives: Objective[]) => {
    const replacement =
      createdObjective &&
      !objectives.some((objective) => objective.id === createdObjective.id)
        ? [createdObjective]
        : [];
    return objectives.flatMap((objective) =>
      objective.id === context.optimisticId ? replacement : [objective],
    );
  };

  for (const queryKey of context.queryKeys) {
    queryClient.setQueryData<ObjectiveListData>(queryKey, (data) => {
      if (Array.isArray(data)) return settleList(data);
      if (!isPaginatedObjectiveList(data)) return data;

      const alreadyCreated = data.pages.some((page) =>
        page.objectives.some(
          (objective) => objective.id === createdObjective?.id,
        ),
      );
      return {
        ...data,
        pages: data.pages.map((page) => ({
          ...page,
          objectives: alreadyCreated
            ? page.objectives.filter(
                (objective) => objective.id !== context.optimisticId,
              )
            : settleList(page.objectives),
        })),
      };
    });
  }
};
