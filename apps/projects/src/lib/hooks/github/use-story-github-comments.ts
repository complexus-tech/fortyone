import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { getStoryGitHubComments } from "@/lib/queries/github/get-story-github-comments";

export const useStoryGitHubComments = (storyId: string, enabled = true) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: githubKeys.storyComments(workspaceSlug, storyId),
    queryFn: () =>
      getStoryGitHubComments(storyId, { session: session!, workspaceSlug }),
    enabled: !!storyId && enabled,
  });
};
