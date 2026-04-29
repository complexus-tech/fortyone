import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { updateGitHubWorkspaceSettingsAction } from "@/lib/actions/github/update-workspace-settings";
import type {
  GitHubIntegration,
  UpdateGitHubWorkspaceSettingsInput,
} from "@/modules/settings/workspace/integrations/github/types";

export const useUpdateGitHubWorkspaceSettings = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = githubKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: (input: UpdateGitHubWorkspaceSettingsInput) =>
      updateGitHubWorkspaceSettingsAction(input, workspaceSlug),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<GitHubIntegration>(queryKey);

      queryClient.setQueryData<GitHubIntegration>(queryKey, (old) => {
        if (!old) return old;
        return {
          ...old,
          settings: {
            ...old.settings,
            ...input,
          },
        };
      });

      return { previous };
    },
    onError: (_err, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous);
      }
      toast.error("GitHub", { description: "Failed to update settings" });
    },
    onSuccess: (res, _input, context) => {
      if (res.error?.message) {
        if (context?.previous) {
          queryClient.setQueryData(queryKey, context.previous);
        }
        toast.error("GitHub", { description: res.error.message });
        return;
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
