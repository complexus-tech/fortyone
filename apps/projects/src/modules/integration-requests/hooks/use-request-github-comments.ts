import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { githubKeys } from "@/constants/keys";
import { getRequestGitHubComments } from "../queries/get-request-github-comments";

export const useRequestGitHubComments = (requestId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: githubKeys.requestComments(workspaceSlug, requestId),
    queryFn: () =>
      getRequestGitHubComments(requestId, { session: session!, workspaceSlug }),
    enabled: Boolean(requestId && session),
  });
};
