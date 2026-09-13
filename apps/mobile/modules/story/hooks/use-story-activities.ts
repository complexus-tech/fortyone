import { useInfiniteQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getStoryActivities } from "../queries/get-activities";

export const useStoryActivitiesInfinite = (id: string, enabled = true) => {
  return useInfiniteQuery({
    enabled,
    queryKey: storyKeys.activitiesInfinite(id),
    queryFn: ({ pageParam, signal }) =>
      getStoryActivities(id, pageParam as number, signal),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
  });
};
