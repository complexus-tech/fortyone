import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { resyncGitHubRepositoriesAction } from "@/lib/actions/github/resync-repositories";

export const useResyncGitHubRepositories = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: () => resyncGitHubRepositoriesAction(workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("GitHub", { description: res.error.message });
        return;
      }
      queryClient.invalidateQueries({
        queryKey: githubKeys.integration(workspaceSlug),
      });
      toast.success("GitHub repositories synced");
    },
  });
};
