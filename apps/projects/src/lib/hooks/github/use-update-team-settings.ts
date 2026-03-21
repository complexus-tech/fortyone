import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { updateGitHubTeamSettingsAction } from "@/lib/actions/github/update-team-settings";
import type { UpdateGitHubTeamSettingsInput } from "@/modules/settings/workspace/integrations/github/types";

export const useUpdateGitHubTeamSettings = (teamId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (input: UpdateGitHubTeamSettingsInput) =>
      updateGitHubTeamSettingsAction(teamId, input, workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("GitHub", { description: res.error.message });
        return;
      }
      queryClient.invalidateQueries({
        queryKey: githubKeys.teamSettings(workspaceSlug, teamId),
      });
    },
  });
};
