import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { GroupStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupStories } from "../queries/get-group-stories";

export const useGroupStories = (params: GroupStoryParams) => {
  const { data: session } = useSession();

  const queryKey = storyKeys.groupStories(params.groupKey, params);

  return useQuery({
    queryKey,
    queryFn: () => getGroupStories(session!, params),
    enabled: Boolean(session && params.groupKey),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
};
