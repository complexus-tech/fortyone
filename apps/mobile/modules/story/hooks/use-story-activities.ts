import { useInfiniteQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getStoryActivities } from "../queries/get-activities";

export const useStoryActivitiesInfinite = (id: string) => {
  return useInfiniteQuery({
    queryKey: storyKeys.activitiesInfinite(id),
    queryFn: ({ pageParam }) => getStoryActivities(id, pageParam as number),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
  });
};
