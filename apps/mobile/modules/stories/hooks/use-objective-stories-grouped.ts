import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getMyStoriesGrouped } from "@/modules/my-work/queries/get-my-stories-grouped";
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
    queryFn: () => getMyStoriesGrouped(params),
    enabled: Boolean(objectiveId),
  });
};
