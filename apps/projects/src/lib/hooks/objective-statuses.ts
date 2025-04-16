import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { objectiveKeys } from "../../modules/objectives/constants";
import { getObjectiveStatuses } from "../../modules/objectives/queries/statuses";

export const useObjectiveStatuses = () => {
  return useQuery({
    queryKey: objectiveKeys.statuses(),
    queryFn: () => getObjectiveStatuses(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
