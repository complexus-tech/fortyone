import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { GroupStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupStories } from "../queries/get-group-stories";
import { buildQueryKey } from "../utils/query-builders";

export const useGroupStories = (params: GroupStoryParams) => {
  const { data: session } = useSession();

  const paramsKey = buildQueryKey([
    params.groupKey,
    params.groupBy,
    params.page,
    params.pageSize,
    params.assignedToMe,
    params.createdByMe,
    params.statusIds?.join(","),
    params.assigneeIds?.join(","),
    params.priorities?.join(","),
    params.teamIds?.join(","),
    params.objectiveId,
    params.sprintIds?.join(","),
  ]);

  const queryKey = [...storyKeys.groupStories(params.groupKey, paramsKey)];

  return useQuery({
    queryKey,
    queryFn: () => getGroupStories(session!, params),
    enabled: Boolean(session && params.groupKey),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
};
