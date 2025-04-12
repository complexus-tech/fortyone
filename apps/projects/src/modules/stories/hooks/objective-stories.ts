import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useObjectiveStories = (objectiveId: string) => {
  return useQuery({
    queryKey: storyKeys.objective(objectiveId),
    queryFn: () => getStories({ objectiveId }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 3,
  });
};
