import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getMyStoriesGrouped } from "../queries/get-my-stories-grouped";
import type { GroupedStoryParams } from "@/modules/stories/types";

export const useMyStoriesGrouped = (
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>
) => {
  const params: GroupedStoryParams = {
    groupBy,
    ...options,
  };

  const queryKey = storyKeys.mineGrouped(params);

  return useQuery({
    queryKey,
    queryFn: () => getMyStoriesGrouped(params),
  });
};
