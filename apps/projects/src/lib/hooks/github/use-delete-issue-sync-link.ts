import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { deleteGitHubIssueSyncLinkAction } from "@/lib/actions/github/delete-issue-sync-link";
import type { GitHubIntegration } from "@/modules/settings/workspace/integrations/github/types";

export const useDeleteGitHubIssueSyncLink = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = githubKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: (linkId: string) =>
      deleteGitHubIssueSyncLinkAction(linkId, workspaceSlug),
    onMutate: async (linkId) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<GitHubIntegration>(queryKey);

      queryClient.setQueryData<GitHubIntegration>(queryKey, (old) => {
        if (!old) return old;
        return {
          ...old,
          issueSyncLinks: old.issueSyncLinks.filter((item) => item.id !== linkId),
        };
      });

      return { previous };
    },
    onError: (_err, _linkId, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous);
      }
      toast.error("GitHub", { description: "Failed to unlink repository" });
    },
    onSuccess: (res, _linkId, context) => {
      if (res.error?.message) {
        if (context?.previous) {
          queryClient.setQueryData(queryKey, context.previous);
        }
        toast.error("GitHub", { description: res.error.message });
        return;
      }
      toast.success("Repository unlinked");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
