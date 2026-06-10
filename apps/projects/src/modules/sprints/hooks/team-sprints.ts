import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getTeamSprints } from "../queries/get-team-sprints";

export const useTeamSprints = (teamId: string, search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();
  return useQuery({
    queryKey: [...sprintKeys.team(workspaceSlug, teamId), normalizedSearch],
    queryFn: () =>
      getTeamSprints(
        teamId,
        { session: session!, workspaceSlug },
        normalizedSearch,
      ),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
