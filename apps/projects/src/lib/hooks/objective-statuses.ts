import { useQuery } from "@tanstack/react-query";
import { objectiveKeys } from "../../modules/objectives/constants";
import { getObjectiveStatuses } from "../../modules/objectives/queries/statuses";

export const useObjectiveStatuses = () => {
  return useQuery({
    queryKey: objectiveKeys.statuses(),
    queryFn: getObjectiveStatuses,
  });
};
