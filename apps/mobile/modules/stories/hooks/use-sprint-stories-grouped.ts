import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getGroupedStories } from "../queries/get-grouped-stories";
import type { GroupedStoryParams } from "../types";

export const useSprintStoriesGrouped = (
  sprintId: string,
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>
) => {
  const params: GroupedStoryParams = {
    groupBy,
    sprintIds: [sprintId],
    ...options,
  };

  const queryKey = storyKeys.sprintGrouped(sprintId, params);

  return useQuery({
    queryKey,
    queryFn: () => getGroupedStories(params),
    enabled: Boolean(sprintId),
    staleTime: 1000 * 60 * 2,
  });
};
