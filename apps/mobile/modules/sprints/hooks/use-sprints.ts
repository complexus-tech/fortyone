import { useQuery } from "@tanstack/react-query";
import { getSprints, getTeamSprints } from "../queries/get-sprints";
import { sprintKeys } from "@/constants/keys";

export const useSprints = () => {
  return useQuery({
    queryKey: sprintKeys.lists(),
    queryFn: getSprints,
  });
};

export const useTeamSprints = (teamId: string) => {
  return useQuery({
    queryKey: sprintKeys.team(teamId),
    queryFn: () => getTeamSprints(teamId),
    enabled: Boolean(teamId),
  });
};
