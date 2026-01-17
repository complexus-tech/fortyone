import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import type { GroupStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupStories } from "../queries/get-group-stories";

export const useGroupStories = (params: GroupStoryParams) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  const queryKey = storyKeys.groupStories(workspaceSlug, params.groupKey, params);

  return useQuery({
    queryKey,
    queryFn: () =>
      getGroupStories({ session: session!, workspaceSlug }, params),
    enabled: Boolean(session && params.groupKey),
    staleTime: 1000 * 60 * 2,
  });
};
