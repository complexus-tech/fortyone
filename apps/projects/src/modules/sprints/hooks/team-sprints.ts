import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getTeamSprints } from "../queries/get-team-sprints";

export const useTeamSprints = (teamId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: sprintKeys.team(workspaceSlug, teamId),
    queryFn: () => getTeamSprints(teamId, { session: session!, workspaceSlug }),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
