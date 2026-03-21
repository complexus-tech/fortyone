import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { updateGitHubWorkspaceSettingsAction } from "@/lib/actions/github/update-workspace-settings";
import type { UpdateGitHubWorkspaceSettingsInput } from "@/modules/settings/workspace/integrations/github/types";

export const useUpdateGitHubWorkspaceSettings = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (input: UpdateGitHubWorkspaceSettingsInput) =>
      updateGitHubWorkspaceSettingsAction(input, workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("GitHub", { description: res.error.message });
        return;
      }
      queryClient.invalidateQueries({
        queryKey: githubKeys.integration(workspaceSlug),
      });
    },
  });
};
