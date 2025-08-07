import { useInfiniteQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStoryComments } from "../queries/get-comments";

export const useStoryCommentsInfinite = (id: string) => {
  const { data: session } = useSession();

  return useInfiniteQuery({
    queryKey: storyKeys.commentsInfinite(id),
    queryFn: ({ pageParam }) => getStoryComments(id, session!, pageParam),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
