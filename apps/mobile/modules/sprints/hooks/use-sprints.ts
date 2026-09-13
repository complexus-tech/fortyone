import { useQuery } from "@tanstack/react-query";
import { sprintKeys } from "@/constants/keys";
import { getSprint, getSprints, getTeamSprints } from "../queries/get-sprints";

export const useSprints = () => {
  return useQuery({
    queryKey: sprintKeys.lists(),
    queryFn: ({ signal }) => getSprints(signal),
  });
};

export const useTeamSprints = (teamId: string) => {
  return useQuery({
    queryKey: sprintKeys.team(teamId),
    queryFn: ({ signal }) => getTeamSprints(teamId, signal),
    enabled: Boolean(teamId),
  });
};

export const useSprint = (sprintId: string) => {
  return useQuery({
    queryKey: sprintKeys.detail(sprintId),
    queryFn: ({ signal }) => getSprint(sprintId, signal),
    enabled: Boolean(sprintId),
  });
};
