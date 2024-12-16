import { storyKeys } from "@/modules/stories/constants";
import { useQuery } from "@tanstack/react-query";
import { getStoryActivities } from "../queries/get-activities";

export const useStoryActivities = (id: string) => {
  return useQuery({
    queryKey: storyKeys.activities(id),
    queryFn: () => getStoryActivities(id),
  });
};
