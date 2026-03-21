import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { deleteGitHubIssueSyncLinkAction } from "@/lib/actions/github/delete-issue-sync-link";

export const useDeleteGitHubIssueSyncLink = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (linkId: string) =>
      deleteGitHubIssueSyncLinkAction(linkId, workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("GitHub", { description: res.error.message });
        return;
      }
      queryClient.invalidateQueries({
        queryKey: githubKeys.integration(workspaceSlug),
      });
      toast.success("Repository unlinked");
    },
  });
};
