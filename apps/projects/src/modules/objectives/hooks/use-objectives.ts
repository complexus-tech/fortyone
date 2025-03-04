import { useQuery } from "@tanstack/react-query";
import { getObjectives } from "../queries/get-objectives";
import { objectiveKeys } from "../constants";

export const useObjectives = () => {
  return useQuery({
    queryKey: objectiveKeys.list(),
    queryFn: getObjectives,
  });
};

export const useTeamObjectives = (teamId: string) => {
  const { data: objectives = [] } = useObjectives();
  return objectives.filter((objective) => objective.teamId === teamId);
};
