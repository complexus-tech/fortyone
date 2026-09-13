import { useQuery } from "@tanstack/react-query";
import { objectiveKeys } from "@/constants/keys";
import {
  getObjective,
  getObjectives,
  getObjectiveStatuses,
  getTeamObjectives,
} from "../queries/get-objectives";

export const useObjectives = () => {
  return useQuery({
    queryKey: objectiveKeys.lists(),
    queryFn: ({ signal }) => getObjectives(signal),
  });
};

export const useTeamObjectives = (teamId: string) => {
  return useQuery({
    queryKey: objectiveKeys.team(teamId),
    queryFn: ({ signal }) => getTeamObjectives(teamId, signal),
    enabled: Boolean(teamId),
  });
};

export const useObjective = (objectiveId: string) => {
  return useQuery({
    queryKey: objectiveKeys.detail(objectiveId),
    queryFn: ({ signal }) => getObjective(objectiveId, signal),
    enabled: Boolean(objectiveId),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
};

export const useObjectiveStatuses = () => {
  return useQuery({
    queryKey: objectiveKeys.statuses(),
    queryFn: ({ signal }) => getObjectiveStatuses(signal),
  });
};
