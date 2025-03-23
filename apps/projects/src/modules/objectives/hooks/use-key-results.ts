import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { objectiveKeys } from "../constants";
import { getKeyResults } from "../queries/get-key-results";

export const useKeyResults = (objectiveId: string) => {
  return useQuery({
    queryKey: objectiveKeys.keyResults(objectiveId),
    queryFn: () => getKeyResults(objectiveId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
