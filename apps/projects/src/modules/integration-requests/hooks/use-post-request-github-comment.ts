import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { githubKeys } from "@/constants/keys";
import { postRequestGitHubCommentAction } from "../actions/post-github-comment";

export const usePostRequestGitHubComment = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: ({ requestId, body }: { requestId: string; body: string }) =>
      postRequestGitHubCommentAction(requestId, { body }, workspaceSlug),
    onError: (error) => {
      toast.error("Failed to post comment to GitHub", {
        description: error.message || "Your comment was not posted",
      });
    },
    onSuccess: (res, { requestId }) => {
      if (res.error?.message) {
        toast.error("Failed to post comment to GitHub", {
          description: res.error.message,
        });
        return;
      }
      toast.success("Comment posted to GitHub");
      queryClient.invalidateQueries({
        queryKey: githubKeys.requestComments(workspaceSlug, requestId),
      });
    },
  });
};
