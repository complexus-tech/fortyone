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

export const useStrategyMap = () => {
  const { session, workspaceSlug } = useStrategyContext();
  return useQuery({
    queryKey: strategyKeys.map(workspaceSlug),
    queryFn: () => getStrategyMap({ session: session!, workspaceSlug }),
    enabled: Boolean(session && workspaceSlug),
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
  const handleError = (error: Error) =>
    toast.error("Strategy could not be updated", {
      description: error.message,
    });

  return {
    updateStrategy: useMutation({
      mutationFn: (data: Parameters<typeof updateStrategy>[0]) =>
        updateStrategy(data, ctx),
      onSuccess: refresh,
      onError: handleError,
    }),
    createPillar: useMutation({
      mutationFn: (data: Parameters<typeof createStrategicPillar>[0]) =>
        createStrategicPillar(data, ctx),
      onSuccess: refresh,
      onError: handleError,
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
      onError: handleError,
    }),
    deletePillar: useMutation({
      mutationFn: (pillarId: string) => deleteStrategicPillar(pillarId, ctx),
      onSuccess: refresh,
      onError: handleError,
    }),
    alignObjective: useMutation({
      mutationFn: ({
        objectiveId,
        pillarId,
      }: {
        objectiveId: string;
        pillarId: string | null;
      }) => alignObjective(objectiveId, pillarId, ctx),
      onSuccess: () => {
        refresh();
        queryClient.invalidateQueries({
          queryKey: objectiveKeys.list(workspaceSlug),
        });
      },
      onError: handleError,
    }),
  };
};
