import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { updateGitHubTeamSettingsAction } from "@/lib/actions/github/update-team-settings";
import type {
  GitHubTeamSettings,
  UpdateGitHubTeamSettingsInput,
} from "@/modules/settings/workspace/integrations/github/types";

export const useUpdateGitHubTeamSettings = (teamId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = githubKeys.teamSettings(workspaceSlug, teamId);

  return useMutation({
    mutationFn: (input: UpdateGitHubTeamSettingsInput) =>
      updateGitHubTeamSettingsAction(teamId, input, workspaceSlug),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<GitHubTeamSettings>(queryKey);

      queryClient.setQueryData<GitHubTeamSettings>(queryKey, (old) => {
        if (!old) return old;
        return {
          ...old,
          rules: input.rules.map((rule) => {
            const existing = old.rules.find(
              (r) => r.eventKey === rule.eventKey,
            );
            return {
              id: existing?.id ?? `${rule.eventKey}-draft`,
              eventKey: rule.eventKey,
              targetStatusId: rule.targetStatusId ?? null,
              baseBranchPattern: rule.baseBranchPattern ?? null,
              isActive: rule.isActive,
            };
          }),
        };
      });

      return { previous };
    },
    onError: (_err, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous);
      }
      toast.error("GitHub", { description: "Failed to update automations" });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("GitHub", { description: res.error.message });
        queryClient.invalidateQueries({ queryKey });
        return;
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
