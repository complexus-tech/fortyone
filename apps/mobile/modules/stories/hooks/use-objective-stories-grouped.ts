import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getGroupedStories } from "../queries/get-grouped-stories";
import type { GroupedStoryParams } from "../types";

export const useObjectiveStoriesGrouped = (
  objectiveId: string,
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>
) => {
  const params: GroupedStoryParams = {
    groupBy,
    objectiveId,
    ...options,
  };

  const queryKey = storyKeys.objectiveGrouped(objectiveId, params);

  return useQuery({
    queryKey,
    queryFn: () => getGroupedStories(params),
    enabled: Boolean(objectiveId),
    staleTime: 1000 * 60 * 2,
  });
};
