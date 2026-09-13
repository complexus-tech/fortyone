import { useInfiniteQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getStoryComments } from "../queries/get-comments";

export const useStoryCommentsInfinite = (id: string, enabled = true) => {
  return useInfiniteQuery({
    enabled,
    queryKey: storyKeys.commentsInfinite(id),
    queryFn: ({ pageParam, signal }) =>
      getStoryComments(id, pageParam as number, signal),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
  });
};
