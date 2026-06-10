import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getObjectives } from "../queries/get-objectives";
import { objectiveKeys } from "../constants";
import { getTeamObjectives } from "../queries/get-team-objectives";

export const useObjectives = (search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();
  return useQuery({
    queryKey: [...objectiveKeys.list(workspaceSlug), normalizedSearch],
    queryFn: () =>
      getObjectives({ session: session!, workspaceSlug }, normalizedSearch),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useTeamObjectives = (teamId: string, search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();
  return useQuery({
    queryKey: [...objectiveKeys.team(workspaceSlug, teamId), normalizedSearch],
    queryFn: () =>
      getTeamObjectives(
        teamId,
        { session: session!, workspaceSlug },
        normalizedSearch,
      ),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
