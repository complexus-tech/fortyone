import { useInfiniteQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getStoryComments } from "../queries/get-comments";

export const useStoryCommentsInfinite = (id: string) => {
  return useInfiniteQuery({
    queryKey: storyKeys.commentsInfinite(id),
    queryFn: ({ pageParam }) => getStoryComments(id, pageParam as number),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
  });
};
