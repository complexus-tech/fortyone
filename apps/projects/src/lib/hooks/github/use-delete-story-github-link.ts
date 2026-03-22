import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { githubKeys } from "@/constants/keys";
import { deleteStoryGitHubLinkAction } from "@/lib/actions/github/delete-story-github-link";

export const useDeleteStoryGitHubLink = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: ({
      storyId,
      linkId,
    }: {
      storyId: string;
      linkId: string;
    }) => deleteStoryGitHubLinkAction(storyId, linkId, workspaceSlug),
    onError: (error) => {
      toast.error("Failed to unlink from GitHub", {
        description: error.message,
      });
    },
    onSuccess: (res, { storyId }) => {
      if (res.error?.message) {
        toast.error("Failed to unlink from GitHub", {
          description: res.error.message,
        });
        return;
      }
      toast.success("GitHub link removed");
      queryClient.invalidateQueries({
        queryKey: githubKeys.storyLinks(workspaceSlug, storyId),
      });
      queryClient.invalidateQueries({
        queryKey: githubKeys.storyComments(workspaceSlug, storyId),
      });
    },
  });
};
