import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import type { GroupedStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupedStories } from "../queries/get-grouped-stories";

export const useGroupedStories = (params: GroupedStoryParams) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  const queryKey = [...storyKeys.grouped(workspaceSlug), params] as const;

  return useQuery({
    queryKey,
    queryFn: () =>
      getGroupedStories({ session: session!, workspaceSlug }, params),
    staleTime: 1000 * 60 * 2,
  });
};
