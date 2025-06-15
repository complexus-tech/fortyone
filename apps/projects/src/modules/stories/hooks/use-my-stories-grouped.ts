import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { GroupedStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupedStories } from "../queries/get-grouped-stories";
import { buildQueryKey } from "../utils/query-builders";

export const useMyStoriesGrouped = (
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>,
) => {
  const { data: session } = useSession();

  const params: GroupedStoryParams = {
    groupBy,
    // assignedToMe: true,
    createdByMe: true,
    storiesPerGroup: 15,
    ...options,
  };

  const paramsKey = buildQueryKey([
    params.groupBy,
    params.assignedToMe,
    params.createdByMe,
    params.storiesPerGroup,
    params.statusIds?.join(","),
    params.priorities?.join(","),
  ]);

  const queryKey = storyKeys.mineGrouped(paramsKey);

  return useQuery({
    queryKey,
    queryFn: () => getGroupedStories(session!, params),
    staleTime: 1000 * 60 * 2,
  });
};
