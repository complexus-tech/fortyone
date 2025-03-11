import { useQuery } from "@tanstack/react-query";
import { getObjectives } from "../queries/get-objectives";
import { objectiveKeys } from "../constants";
import { getTeamObjectives } from "../queries/get-team-objectives";

export const useObjectives = () => {
  return useQuery({
    queryKey: objectiveKeys.list(),
    queryFn: getObjectives,
  });
};

export const useTeamObjectives = (teamId: string) => {
  return useQuery({
    queryKey: objectiveKeys.team(teamId),
    queryFn: () => getTeamObjectives(teamId),
  });
};
