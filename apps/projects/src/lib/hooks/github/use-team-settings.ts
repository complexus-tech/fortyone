import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { getGitHubTeamSettings } from "@/lib/queries/github/get-team-settings";

export const useGitHubTeamSettings = (teamId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: githubKeys.teamSettings(workspaceSlug, teamId),
    queryFn: () =>
      getGitHubTeamSettings(teamId, { session: session!, workspaceSlug }),
  });
};
