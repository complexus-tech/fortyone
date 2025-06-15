import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { GroupedStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupedStories } from "../queries/get-grouped-stories";
import { buildQueryKey } from "../utils/query-builders";

export const useGroupedStories = (params: GroupedStoryParams) => {
  const { data: session } = useSession();

  const paramsKey = buildQueryKey([
    params.groupBy,
    params.teamIds?.join(","),
    params.assignedToMe,
    params.createdByMe,
    params.storiesPerGroup,
    params.statusIds?.join(","),
    params.assigneeIds?.join(","),
    params.priorities?.join(","),
    params.objectiveId,
    params.sprintIds?.join(","),
  ]);

  const queryKey = [...storyKeys.grouped(), paramsKey || "default"] as const;

  return useQuery({
    queryKey,
    queryFn: () => getGroupedStories(session!, params),
    enabled: Boolean(session),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
};
