import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { objectiveKeys } from "@/modules/objectives/constants";
import {
  alignObjective,
  createStrategicPillar,
  deleteStrategicPillar,
  getStrategyMap,
  updateStrategicPillar,
  updateStrategy,
} from "./api";
import { alignObjectiveInStrategy } from "./strategy-cache";
import type { StrategyMap } from "./types";

export const strategyKeys = {
  map: (workspaceSlug: string) => ["strategy-map", workspaceSlug] as const,
};

const useStrategyContext = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return { session, workspaceSlug };
};

const handleStrategyError = (error: Error) =>
  toast.error("Strategy could not be updated", {
    description: error.message,
  });

export const useStrategyMap = () => {
  const { session, workspaceSlug } = useStrategyContext();
  return useQuery({
    queryKey: strategyKeys.map(workspaceSlug),
    queryFn: () => getStrategyMap({ session: session!, workspaceSlug }),
    enabled: Boolean(session && workspaceSlug),
  });
};

export const useAlignObjectiveMutation = () => {
  const queryClient = useQueryClient();
  const { session, workspaceSlug } = useStrategyContext();
  const ctx = { session: session!, workspaceSlug };
  const strategyKey = strategyKeys.map(workspaceSlug);

  return useMutation({
    mutationFn: ({
      objectiveId,
      pillarId,
    }: {
      objectiveId: string;
      pillarId: string | null;
    }) => alignObjective(objectiveId, pillarId, ctx),
    onMutate: async ({ objectiveId, pillarId }) => {
      await queryClient.cancelQueries({ queryKey: strategyKey });
      const previousStrategy =
        queryClient.getQueryData<StrategyMap>(strategyKey);

      queryClient.setQueryData<StrategyMap>(strategyKey, (strategy) => {
        if (!strategy) return strategy;
        return alignObjectiveInStrategy(strategy, objectiveId, pillarId);
      });

      return { previousStrategy };
    },
    onError: (error, _variables, context) => {
      if (context?.previousStrategy) {
        queryClient.setQueryData(strategyKey, context.previousStrategy);
      }
      handleStrategyError(error);
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: strategyKey });
      void queryClient.invalidateQueries({
        queryKey: objectiveKeys.list(workspaceSlug),
      });
    },
  });
};

export const useStrategyMutations = () => {
  const queryClient = useQueryClient();
  const { session, workspaceSlug } = useStrategyContext();
  const ctx = { session: session!, workspaceSlug };
  const strategyKey = strategyKeys.map(workspaceSlug);
  const refresh = () =>
    queryClient.invalidateQueries({
      queryKey: strategyKey,
    });
  const optimisticallyUpdateStrategy = async (
    update: (strategy: StrategyMap) => StrategyMap,
  ) => {
    await queryClient.cancelQueries({ queryKey: strategyKey });
    const previousStrategy = queryClient.getQueryData<StrategyMap>(strategyKey);
    queryClient.setQueryData<StrategyMap>(strategyKey, (strategy) =>
      strategy ? update(strategy) : strategy,
    );
    return { previousStrategy };
  };
  const handleOptimisticError = (
    error: Error,
    context?: { previousStrategy?: StrategyMap },
  ) => {
    if (context?.previousStrategy) {
      queryClient.setQueryData(strategyKey, context.previousStrategy);
    }
    handleStrategyError(error);
  };
  const alignObjectiveMutation = useAlignObjectiveMutation();

  return {
    updateStrategy: useMutation({
      mutationFn: (data: Parameters<typeof updateStrategy>[0]) =>
        updateStrategy(data, ctx),
      onMutate: (data) =>
        optimisticallyUpdateStrategy((strategy) => ({ ...strategy, ...data })),
      onError: (error, _variables, context) => {
        handleOptimisticError(error, context);
      },
      onSettled: refresh,
    }),
    createPillar: useMutation({
      mutationFn: (data: Parameters<typeof createStrategicPillar>[0]) =>
        createStrategicPillar(data, ctx),
      onMutate: async (data) => {
        const optimisticPillarId = `optimistic-${crypto.randomUUID()}`;
        const context = await optimisticallyUpdateStrategy((strategy) => ({
          ...strategy,
          pillars: [
            ...strategy.pillars,
            {
              ...data,
              id: optimisticPillarId,
              objectiveIds: [],
            },
          ],
        }));
        return { ...context, optimisticPillarId };
      },
      onSuccess: (response, _variables, context) => {
        const createdPillar = response.data;
        if (!createdPillar) return;

        queryClient.setQueryData<StrategyMap>(strategyKey, (strategy) =>
          strategy
            ? {
                ...strategy,
                pillars: strategy.pillars.map((pillar) =>
                  pillar.id === context.optimisticPillarId
                    ? createdPillar
                    : pillar,
                ),
              }
            : strategy,
        );
      },
      onError: (error, _variables, context) => {
        handleOptimisticError(error, context);
      },
      onSettled: refresh,
    }),
    updatePillar: useMutation({
      mutationFn: ({
        pillarId,
        data,
      }: {
        pillarId: string;
        data: Parameters<typeof updateStrategicPillar>[1];
      }) => updateStrategicPillar(pillarId, data, ctx),
      onMutate: ({ pillarId, data }) =>
        optimisticallyUpdateStrategy((strategy) => ({
          ...strategy,
          pillars: strategy.pillars.map((pillar) =>
            pillar.id === pillarId
              ? {
                  ...pillar,
                  description:
                    data.description === undefined
                      ? pillar.description
                      : data.description,
                  name: data.name ?? pillar.name,
                  orderIndex: data.orderIndex ?? pillar.orderIndex,
                }
              : pillar,
          ),
        })),
      onError: (error, _variables, context) => {
        handleOptimisticError(error, context);
      },
      onSettled: refresh,
    }),
    deletePillar: useMutation({
      mutationFn: (pillarId: string) => deleteStrategicPillar(pillarId, ctx),
      onMutate: (pillarId) =>
        optimisticallyUpdateStrategy((strategy) => ({
          ...strategy,
          pillars: strategy.pillars.filter((pillar) => pillar.id !== pillarId),
        })),
      onError: (error, _variables, context) => {
        handleOptimisticError(error, context);
      },
      onSettled: refresh,
    }),
    alignObjective: alignObjectiveMutation,
  };
};
