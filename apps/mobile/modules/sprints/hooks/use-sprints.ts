import { useQuery } from "@tanstack/react-query";
import { sprintKeys } from "@/constants/keys";
import { getSprint, getSprints, getTeamSprints } from "../queries/get-sprints";

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

export const useSprint = (sprintId: string) => {
  return useQuery({
    queryKey: sprintKeys.detail(sprintId),
    queryFn: () => getSprint(sprintId),
    enabled: Boolean(sprintId),
  });
};
