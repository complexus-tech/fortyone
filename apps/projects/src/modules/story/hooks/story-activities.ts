import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStoryActivities } from "../queries/get-activities";

export const useStoryActivities = (id: string) => {
  return useQuery({
    queryKey: storyKeys.activities(id),
    queryFn: () => getStoryActivities(id),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
