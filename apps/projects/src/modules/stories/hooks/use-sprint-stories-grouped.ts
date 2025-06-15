import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { GroupedStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupedStories } from "../queries/get-grouped-stories";
import { buildQueryKey } from "../utils/query-builders";

export const useSprintStoriesGrouped = (
  sprintId: string,
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>,
) => {
  const { data: session } = useSession();

  const params: GroupedStoryParams = {
    groupBy,
    sprintIds: [sprintId],
    storiesPerGroup: 30,
    ...options,
  };

  const paramsKey = buildQueryKey([
    params.groupBy,
    params.sprintIds?.join(","),
    params.assignedToMe,
    params.createdByMe,
    params.storiesPerGroup,
    params.teamIds?.join(","),
    params.statusIds?.join(","),
    params.priorities?.join(","),
  ]);

  const queryKey = storyKeys.sprintGrouped(sprintId, paramsKey);

  return useQuery({
    queryKey,
    queryFn: () => getGroupedStories(session!, params),
    enabled: Boolean(sprintId),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
};
