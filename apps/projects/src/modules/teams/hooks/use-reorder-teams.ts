import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import { reorderTeamsAction } from "../actions/reorder-teams";
import type { Team } from "../types";
import { reorderCachedTeams } from "./team-order-cache";

export const useReorderTeamsMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();
  const toastId = "reorder-teams";

  const mutation = useMutation({
    mutationFn: async (data: Parameters<typeof reorderTeamsAction>[0]) => {
      const response = await reorderTeamsAction(data, workspaceSlug);

      if (response.error?.message) {
        throw new Error(response.error.message);
      }

      return response;
    },
    onMutate: async (data) => {
      await queryClient.cancelQueries({
        queryKey: teamKeys.lists(workspaceSlug),
      });

      const previousTeams = queryClient.getQueryData<Team[]>(
        teamKeys.lists(workspaceSlug),
      );
      const previousJoinedTeams = queryClient.getQueryData<Team[]>(
        teamKeys.joined(workspaceSlug),
      );

      queryClient.setQueryData<Team[]>(teamKeys.lists(workspaceSlug), (teams) =>
        reorderCachedTeams(teams, data.teamIds),
      );
      queryClient.setQueryData<Team[]>(
        teamKeys.joined(workspaceSlug),
        (teams) => reorderCachedTeams(teams, data.teamIds),
      );

      return { previousJoinedTeams, previousTeams };
    },

    onError: (error, variables, context) => {
      if (context?.previousTeams) {
        queryClient.setQueryData(
          teamKeys.lists(workspaceSlug),
          context.previousTeams,
        );
      }
      if (context?.previousJoinedTeams) {
        queryClient.setQueryData(
          teamKeys.joined(workspaceSlug),
          context.previousJoinedTeams,
        );
      }
      toast.error("Failed to reorder teams", {
        description:
          error.message || "An error occurred while reordering teams",
        id: toastId,
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (_response, variables) => {
      analytics.track("teams_reordered", {
        newOrder: variables.teamIds,
      });

      queryClient.invalidateQueries({
        queryKey: teamKeys.lists(workspaceSlug),
      });
    },
  });

  return mutation;
};
