import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getGroupedStories } from "../queries/get-grouped-stories";
import type { GroupedStoryParams } from "../types";

export const useTeamStoriesGrouped = (
  teamId: string,
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>,
  enabled = true,
) => {
  const params: GroupedStoryParams = {
    groupBy,
    teamIds: [teamId],
    ...options,
  };

  const queryKey = storyKeys.teamGrouped(teamId, params);

  return useQuery({
    queryKey,
    queryFn: ({ signal }) => getGroupedStories(params, signal),
    enabled: Boolean(teamId) && enabled,
    staleTime: 1000 * 60 * 2,
  });
};
