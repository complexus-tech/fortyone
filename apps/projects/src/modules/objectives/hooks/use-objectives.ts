import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getObjectives } from "../queries/get-objectives";
import { objectiveKeys } from "../constants";
import { getTeamObjectives } from "../queries/get-team-objectives";

export const useObjectives = () => {
  return useQuery({
    queryKey: objectiveKeys.list(),
    queryFn: getObjectives,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useTeamObjectives = (teamId: string) => {
  return useQuery({
    queryKey: objectiveKeys.team(teamId),
    queryFn: () => getTeamObjectives(teamId),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
