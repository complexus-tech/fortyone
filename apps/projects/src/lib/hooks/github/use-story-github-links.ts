import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { getStoryGitHubLinks } from "@/lib/queries/github/get-story-github-links";

export const useStoryGitHubLinks = (storyId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: githubKeys.storyLinks(workspaceSlug, storyId),
    queryFn: () =>
      getStoryGitHubLinks(storyId, { session: session!, workspaceSlug }),
    enabled: !!storyId,
  });
};
