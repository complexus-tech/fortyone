import { useQuery } from "@tanstack/react-query";
import { sprintKeys } from "@/constants/keys";
import { getTeamSprints } from "../queries/get-team-sprints";

export const useTeamSprints = (teamId: string) => {
  return useQuery({
    queryKey: sprintKeys.team(teamId),
    queryFn: () => getTeamSprints(teamId),
    enabled: Boolean(teamId),
  });
};
