import { useQuery } from "@tanstack/react-query";
import { objectiveKeys } from "../constants";
import { getKeyResults } from "../queries/get-key-results";

export const useKeyResults = (objectiveId: string) => {
  return useQuery({
    queryKey: objectiveKeys.keyResults(objectiveId),
    queryFn: () => getKeyResults(objectiveId),
  });
};
