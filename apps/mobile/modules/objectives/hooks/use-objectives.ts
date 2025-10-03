import { useQuery } from "@tanstack/react-query";
import {
  getObjectives,
  getTeamObjectives,
  getObjectiveStatuses,
} from "../queries/get-objectives";
import { objectiveKeys } from "@/constants/keys";

export const useObjectives = () => {
  return useQuery({
    queryKey: objectiveKeys.lists(),
    queryFn: getObjectives,
  });
};

export const useTeamObjectives = (teamId: string) => {
  return useQuery({
    queryKey: objectiveKeys.team(teamId),
    queryFn: () => getTeamObjectives(teamId),
    enabled: Boolean(teamId),
  });
};

export const useObjectiveStatuses = () => {
  return useQuery({
    queryKey: objectiveKeys.statuses(),
    queryFn: getObjectiveStatuses,
  });
};
