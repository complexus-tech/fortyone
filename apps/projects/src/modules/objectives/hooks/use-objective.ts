import { useQuery } from "@tanstack/react-query";
import { objectiveKeys } from "../constants";
import { getObjective } from "../queries/get-objective";

export const useObjective = (objectiveId: string) => {
  return useQuery({
    queryKey: objectiveKeys.objective(objectiveId),
    queryFn: () => getObjective(objectiveId),
  });
};
