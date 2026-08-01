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

  return useMutation({
    mutationFn: ({
      objectiveId,
      pillarId,
    }: {
      objectiveId: string;
      pillarId: string | null;
    }) => alignObjective(objectiveId, pillarId, ctx),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: strategyKeys.map(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.list(workspaceSlug),
      });
    },
    onError: handleStrategyError,
  });
};

export const useStrategyMutations = () => {
  const queryClient = useQueryClient();
  const { session, workspaceSlug } = useStrategyContext();
  const ctx = { session: session!, workspaceSlug };
  const refresh = () =>
    queryClient.invalidateQueries({
      queryKey: strategyKeys.map(workspaceSlug),
    });
  const alignObjectiveMutation = useAlignObjectiveMutation();

  return {
    updateStrategy: useMutation({
      mutationFn: (data: Parameters<typeof updateStrategy>[0]) =>
        updateStrategy(data, ctx),
      onSuccess: refresh,
      onError: handleStrategyError,
    }),
    createPillar: useMutation({
      mutationFn: (data: Parameters<typeof createStrategicPillar>[0]) =>
        createStrategicPillar(data, ctx),
      onSuccess: refresh,
      onError: handleStrategyError,
    }),
    updatePillar: useMutation({
      mutationFn: ({
        pillarId,
        data,
      }: {
        pillarId: string;
        data: Parameters<typeof updateStrategicPillar>[1];
      }) => updateStrategicPillar(pillarId, data, ctx),
      onSuccess: refresh,
      onError: handleStrategyError,
    }),
    deletePillar: useMutation({
      mutationFn: (pillarId: string) => deleteStrategicPillar(pillarId, ctx),
      onSuccess: refresh,
      onError: handleStrategyError,
    }),
    alignObjective: alignObjectiveMutation,
  };
};
