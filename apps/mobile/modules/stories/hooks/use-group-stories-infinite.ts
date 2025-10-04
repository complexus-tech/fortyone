import { useInfiniteQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getGroupStories } from "../queries/get-group-stories";
import type { GroupStoryParams, StoryGroup } from "../types";

export const useGroupStoriesInfinite = (
  params: GroupStoryParams,
  initialGroup: StoryGroup
) => {
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
        orderBy: params.orderBy || "created",
        orderDirection: params.orderDirection || "desc",
      },
    ],
    pageParams: [1],
  };

  return useInfiniteQuery({
    queryKey: storyKeys.group(params.groupKey, params),
    queryFn: ({ pageParam }) => getGroupStories({ ...params, page: pageParam }),
    getNextPageParam: (lastPage) =>
      lastPage?.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: initialGroup.nextPage,
    initialData,
  });
};
