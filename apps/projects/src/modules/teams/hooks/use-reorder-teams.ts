import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import { reorderTeamsAction } from "../actions/reorder-teams";
import type { Team } from "../types";

export const useReorderTeamsMutation = () => {
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();
  const toastId = "reorder-teams";

  const mutation = useMutation({
    mutationFn: reorderTeamsAction,
    onMutate: (data) => {
      const previousTeams = queryClient.getQueryData<Team[]>(teamKeys.lists());
      // reorder the teams
      const reorderedTeams = data.teamIds.map((id) => {
        return previousTeams?.find((t) => t.id === id);
      });
      queryClient.setQueryData(teamKeys.lists(), reorderedTeams);

      return { previousTeams };
    },

    onError: (error, variables, context) => {
      queryClient.setQueryData(teamKeys.lists(), context?.previousTeams);
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

    onSuccess: (res, variables) => {
      if (res.error?.message) {
        toast.error("Failed to reorder teams", {
          description: res.error.message,
          id: toastId,
        });
        throw new Error(res.error.message);
      }

      analytics.track("teams_reordered", {
        newOrder: variables.teamIds,
      });

      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });

  return mutation;
};
