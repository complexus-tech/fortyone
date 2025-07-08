import { useInfiniteQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getGroupStories } from "../queries/get-group-stories";
import { storyKeys } from "../constants";
import type { GroupStoryParams, StoryGroup } from "../types";

export const useGroupStoriesInfinite = (
  params: GroupStoryParams,
  initialGroup: StoryGroup,
) => {
  const { data: session } = useSession();

  const initialData = {
    pages: [
      {
        groupKey: initialGroup.key,
        stories: initialGroup.stories,
        pagination: {
          page: 1,
          pageSize: initialGroup.loadedCount,
          hasMore: initialGroup.hasMore,
          nextPage: initialGroup.nextPage,
        },
        filters: {},
      },
    ],
    pageParams: [1],
  };

  return useInfiniteQuery({
    queryKey: storyKeys.groupStories(params.groupKey, params),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    queryFn: ({ pageParam }) =>
      getGroupStories(session!, { ...params, page: pageParam }),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: initialGroup.nextPage,
    initialData,
  });
};
