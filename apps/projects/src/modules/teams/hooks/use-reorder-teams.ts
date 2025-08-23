import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import { reorderTeamsAction } from "../actions/reorder-teams";

/**
 * Hook for reordering teams
 *
 * @example
 * ```tsx
 * const reorderTeams = useReorderTeamsMutation();
 *
 * // Reorder teams by their IDs in the desired order
 * reorderTeams.mutate({
 *   teamIds: ["team1", "team2", "team3"]
 * });
 * ```
 */
export const useReorderTeamsMutation = () => {
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();
  const toastId = "reorder-teams";

  const mutation = useMutation({
    mutationFn: reorderTeamsAction,
    onMutate: (data) => {
      const previousTeams = queryClient.getQueryData(teamKeys.lists());

      toast.loading("Reordering teams...", {
        description: "Please wait while we update the team order",
        id: toastId,
      });

      return { previousTeams, teamCount: data.teamIds.length };
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

    onSuccess: (res, variables, context) => {
      if (res.error?.message) {
        toast.error("Failed to reorder teams", {
          description: res.error.message,
          id: toastId,
        });
        throw new Error(res.error.message);
      }

      // Track team reordering
      analytics.track("teams_reordered", {
        teamCount: context.teamCount || 0,
      });

      toast.success("Success", {
        description: "Teams reordered successfully",
        id: toastId,
      });

      // Invalidate teams query to refresh the list
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });

  return mutation;
};
