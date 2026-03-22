import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { githubKeys } from "@/constants/keys";
import { postGitHubCommentAction } from "@/lib/actions/github/post-github-comment";

export const usePostGitHubComment = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: ({
      storyId,
      body,
    }: {
      storyId: string;
      body: string;
    }) => postGitHubCommentAction(storyId, { body }, workspaceSlug),
    onError: (error) => {
      toast.error("Failed to post comment to GitHub", {
        description: error.message || "Your comment was not posted",
      });
    },
    onSuccess: (res, { storyId }) => {
      if (res.error?.message) {
        toast.error("Failed to post comment to GitHub", {
          description: res.error.message,
        });
        return;
      }
      toast.success("Comment posted to GitHub");
      queryClient.invalidateQueries({
        queryKey: githubKeys.storyComments(workspaceSlug, storyId),
      });
    },
  });
};
