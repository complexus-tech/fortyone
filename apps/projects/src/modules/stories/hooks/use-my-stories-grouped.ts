import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { GroupedStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupedStories } from "../queries/get-grouped-stories";

export const useMyStoriesGrouped = (
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>,
) => {
  const { data: session } = useSession();

  const params: GroupedStoryParams = {
    groupBy,
    ...options,
  };

  const queryKey = storyKeys.mineGrouped(params);

  return useQuery({
    queryKey,
    queryFn: () => getGroupedStories(session!, params),
    staleTime: 1000 * 60 * 2,
  });
};
