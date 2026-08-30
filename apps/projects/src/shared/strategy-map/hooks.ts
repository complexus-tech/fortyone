import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { objectiveKeys } from "@/shared/objectives/keys";
import { alignObjective, getStrategyMap } from "./api";
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

type UseAlignObjectiveMutationOptions = {
  onAlignmentSettled?: () => void;
};

export const useAlignObjectiveMutation = ({
  onAlignmentSettled,
}: UseAlignObjectiveMutationOptions = {}) => {
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
      onAlignmentSettled?.();
    },
  });
};
