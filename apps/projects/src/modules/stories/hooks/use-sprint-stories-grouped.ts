import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import type { GroupedStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupedStories } from "../queries/get-grouped-stories";

export const useSprintStoriesGrouped = (
  sprintId: string,
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  const params: GroupedStoryParams = {
    groupBy,
    sprintIds: [sprintId],
    ...options,
  };

  const queryKey = storyKeys.sprintGrouped(workspaceSlug, sprintId, params);

  return useQuery({
    queryKey,
    queryFn: () =>
      getGroupedStories({ session: session!, workspaceSlug }, params),
    enabled: Boolean(sprintId),
    staleTime: 1000 * 60 * 2,
  });
};
