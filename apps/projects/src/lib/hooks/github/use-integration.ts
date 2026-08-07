import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { getGitHubIntegration } from "@/lib/queries/github/get-integration";

export const useGitHubIntegration = (
  { enabled = true }: { enabled?: boolean } = {},
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    enabled: enabled && Boolean(session && workspaceSlug),
    queryKey: githubKeys.integration(workspaceSlug),
    queryFn: () => getGitHubIntegration({ session: session!, workspaceSlug }),
  });
};
