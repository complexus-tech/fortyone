import { useQuery } from "@tanstack/react-query";
import { sprintKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getTeamSprints } from "../queries/get-team-sprints";

export const useTeamSprints = (teamId: string) => {
  return useQuery({
    queryKey: sprintKeys.team(teamId),
    queryFn: () => getTeamSprints(teamId),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
