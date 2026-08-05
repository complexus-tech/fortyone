import type { InfiniteData } from "@tanstack/react-query";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { analyticsKeys } from "@/constants/keys";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { createKeyResults } from "../actions/create-key-results";
import { objectiveKeys } from "../constants";
import type {
  KeyResult,
  NewKeyResult,
  Objective,
  ObjectivesPage,
} from "../types";

type CreateKeyResultsVariables = {
  objectiveId: string;
  keyResults: NewKeyResult[];
};

type ObjectiveQueryData =
  | Objective
  | Objective[]
  | ObjectivesPage
  | InfiniteData<ObjectivesPage>;

const isObjective = (value: unknown): value is Objective =>
  Boolean(
    value &&
      typeof value === "object" &&
      typeof Reflect.get(value, "id") === "string" &&
      typeof Reflect.get(value, "keyResultCount") === "number",
  );

const updateObjectiveCount = (
  data: ObjectiveQueryData | undefined,
  objectiveId: string,
  increment: number,
): ObjectiveQueryData | undefined => {
  if (!data) return data;
  if (Array.isArray(data)) {
    return data.map((objective) =>
      objective.id === objectiveId
        ? {
            ...objective,
            keyResultCount: objective.keyResultCount + increment,
          }
        : objective,
    );
  }
  if (isObjective(data)) {
    return data.id === objectiveId
      ? { ...data, keyResultCount: data.keyResultCount + increment }
      : data;
  }
  if ("objectives" in data) {
    return {
      ...data,
      objectives: updateObjectiveCount(
        data.objectives,
        objectiveId,
        increment,
      ) as Objective[],
    };
  }
  return {
    ...data,
    pages: data.pages.map(
      (page) =>
        updateObjectiveCount(page, objectiveId, increment) as ObjectivesPage,
    ),
  };
};

export const useCreateKeyResultsMutation = () => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  return useMutation({
    mutationFn: async ({
      objectiveId,
      keyResults,
    }: CreateKeyResultsVariables) => {
      const response = await createKeyResults(
        objectiveId,
        keyResults,
        workspaceSlug,
      );
      if (response.error?.message) throw new Error(response.error.message);
      if (!response.data || response.data.length !== keyResults.length) {
        throw new Error("The key-result batch was not created completely.");
      }
      return response.data;
    },
    onMutate: async ({ objectiveId, keyResults }) => {
      const keyResultsQuery = objectiveKeys.keyResults(
        workspaceSlug,
        objectiveId,
      );
      await queryClient.cancelQueries({ queryKey: keyResultsQuery });
      const previousKeyResults =
        queryClient.getQueryData<KeyResult[]>(keyResultsQuery);
      const optimisticIds = keyResults.map(
        () => `optimistic:${crypto.randomUUID()}`,
      );
      const now = new Date().toISOString();
      const optimisticKeyResults = keyResults.map<KeyResult>(
        (keyResult, index) => ({
          ...keyResult,
          id: optimisticIds[index],
          objectiveId,
          sequenceId: 0,
          lead: keyResult.lead ?? null,
          contributors: keyResult.contributors ?? [],
          createdAt: now,
          updatedAt: now,
          createdBy: session?.user.id ?? "",
        }),
      );
      queryClient.setQueryData<KeyResult[]>(keyResultsQuery, (current = []) => [
        ...optimisticKeyResults,
        ...current,
      ]);
      return { optimisticIds, previousKeyResults };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, variables.objectiveId),
        context?.previousKeyResults ?? [],
      );
      toast.error("Key results could not be created", {
        description: error.message,
      });
    },
    onSuccess: (createdKeyResults, variables, context) => {
      const optimisticIds = new Set(context.optimisticIds);
      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, variables.objectiveId),
        (current = []) => [
          ...createdKeyResults,
          ...current.filter(({ id }) => !optimisticIds.has(id)),
        ],
      );
      queryClient.setQueriesData<ObjectiveQueryData>(
        {
          predicate: ({ queryKey }) =>
            queryKey[0] === "objectives" &&
            queryKey[1] === workspaceSlug &&
            queryKey[2] === "list",
        },
        (data) =>
          updateObjectiveCount(
            data,
            variables.objectiveId,
            createdKeyResults.length,
          ),
      );
      createdKeyResults.forEach((keyResult) => {
        analytics.track("key_result_created", {
          keyResultId: keyResult.id,
          objectiveId: keyResult.objectiveId,
          measurementType: keyResult.measurementType,
        });
      });
      toast.success(
        `${createdKeyResults.length} key result${createdKeyResults.length === 1 ? "" : "s"} created`,
      );
    },
    onSettled: (_data, _error, variables) => {
      void Promise.all([
        queryClient.invalidateQueries({
          queryKey: objectiveKeys.keyResults(
            workspaceSlug,
            variables.objectiveId,
          ),
        }),
        queryClient.invalidateQueries({
          queryKey: objectiveKeys.list(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: objectiveKeys.activitiesInfinite(
            workspaceSlug,
            variables.objectiveId,
          ),
        }),
        queryClient.invalidateQueries({
          queryKey: ["key-results", workspaceSlug],
        }),
        queryClient.invalidateQueries({
          queryKey: ["strategy-map", workspaceSlug],
        }),
        queryClient.invalidateQueries({
          queryKey: analyticsKeys.all(workspaceSlug),
        }),
      ]);
    },
  });
};
