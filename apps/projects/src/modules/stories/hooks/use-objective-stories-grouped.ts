import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { GroupedStoryParams } from "../types/grouped";
import { storyKeys } from "../constants";
import { getGroupedStories } from "../queries/get-grouped-stories";
import { buildQueryKey } from "../utils/query-builders";

export const useObjectiveStoriesGrouped = (
  objectiveId: string,
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>,
) => {
  const { data: session } = useSession();

  const params: GroupedStoryParams = {
    groupBy,
    objectiveId,
    storiesPerGroup: 10,
    ...options,
  };

  const paramsKey = buildQueryKey([
    params.groupBy,
    params.objectiveId,
    params.assignedToMe,
    params.createdByMe,
    params.storiesPerGroup,
    params.teamIds?.join(","),
    params.statusIds?.join(","),
    params.priorities?.join(","),
  ]);

  const queryKey = storyKeys.objectiveGrouped(objectiveId, paramsKey);

  return useQuery({
    queryKey,
    queryFn: () => getGroupedStories(session!, params),
    enabled: Boolean(session && objectiveId),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
};
