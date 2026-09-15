import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath, useAnalytics } from "@/hooks";
import {
  analyticsKeys,
  calendarKeys,
  developerKeys,
  feedbackKeys,
  githubKeys,
  integrationRequestKeys,
  keyResultKeys,
  labelKeys,
  memberKeys,
  notificationKeys,
  sprintKeys,
  statusKeys,
  teamKeys,
} from "@/constants/keys";
import { storyKeys } from "@/modules/stories/constants";
import { documentKeys } from "@/shared/documents/keys";
import { objectiveKeys } from "@/shared/objectives/keys";
import { deleteTeamAction } from "../actions/delete-team";

export const useDeleteTeamMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();
  const toastId = "delete-team";

  const mutation = useMutation({
    mutationFn: async (id: string) => {
      const response = await deleteTeamAction(id, workspaceSlug);
      if (response.error) {
        throw new Error(response.error.message || "Failed to delete team");
      }
      return response;
    },
    onMutate: async () => {
      await queryClient.cancelQueries({
        queryKey: teamKeys.lists(workspaceSlug),
      });

      toast.loading("Deleting team...", {
        description: "Please wait while we delete the team",
        id: toastId,
      });
    },
    onError: (error) => {
      toast.dismiss(toastId);
      toast.error("Failed to delete team", {
        description: error.message || "Failed to delete team",
        id: toastId,
      });
    },
    onSuccess: (_response, teamId) => {
      const deletedTeamKeys = [
        teamKeys.detail(workspaceSlug, teamId),
        teamKeys.settings(workspaceSlug, teamId),
        memberKeys.team(workspaceSlug, teamId),
        statusKeys.team(workspaceSlug, teamId),
        labelKeys.team(workspaceSlug, teamId),
        sprintKeys.team(workspaceSlug, teamId),
        storyKeys.team(workspaceSlug, teamId),
        objectiveKeys.team(workspaceSlug, teamId),
        githubKeys.teamSettings(workspaceSlug, teamId),
      ];
      for (const queryKey of deletedTeamKeys) {
        queryClient.removeQueries({ queryKey });
      }

      // Cascades also change workspace-wide lists, totals, and linked data.
      const affectedQueryKeys = [
        teamKeys.lists(workspaceSlug),
        storyKeys.all(workspaceSlug),
        storyKeys.total(workspaceSlug),
        objectiveKeys.all(workspaceSlug),
        keyResultKeys.all(workspaceSlug),
        sprintKeys.all(workspaceSlug),
        statusKeys.all(workspaceSlug),
        labelKeys.all(workspaceSlug),
        feedbackKeys.all(workspaceSlug),
        integrationRequestKeys.all(workspaceSlug),
        documentKeys.all(workspaceSlug),
        notificationKeys.all(workspaceSlug),
        calendarKeys.schedules(workspaceSlug),
        calendarKeys.events(workspaceSlug),
        analyticsKeys.all(workspaceSlug),
        developerKeys.personalTokens(workspaceSlug),
        developerKeys.serviceAccounts(workspaceSlug),
      ];
      for (const queryKey of affectedQueryKeys) {
        void queryClient.invalidateQueries({ queryKey });
      }

      // Track team deletion
      analytics.track("team_deleted", {
        teamId,
      });
      toast.success("Success", {
        description: "Team deleted successfully",
        id: toastId,
      });
    },
  });

  return mutation;
};
