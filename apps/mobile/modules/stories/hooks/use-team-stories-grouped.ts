import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getMyStoriesGrouped } from "@/modules/my-work/queries/get-my-stories-grouped";
import type { GroupedStoryParams } from "../types";

export const useTeamStoriesGrouped = (
  teamId: string,
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>
) => {
  const params: GroupedStoryParams = {
    groupBy,
    teamIds: [teamId],
    ...options,
  };

  const queryKey = storyKeys.teamGrouped(teamId, params);

  return useQuery({
    queryKey,
    queryFn: () => getMyStoriesGrouped(params),
    enabled: Boolean(teamId),
  });
};
