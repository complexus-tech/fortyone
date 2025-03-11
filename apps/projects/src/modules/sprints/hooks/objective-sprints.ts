import { useQuery } from "@tanstack/react-query";
import { sprintKeys } from "@/constants/keys";
import { getObjectiveSprints } from "../queries/get-objective-sprints";

export const useObjectiveSprints = (objectiveId: string) => {
  return useQuery({
    queryKey: sprintKeys.objective(objectiveId),
    queryFn: () => getObjectiveSprints(objectiveId),
    enabled: Boolean(objectiveId),
  });
};
