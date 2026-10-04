import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import type { GroupedStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupedStories } from "../queries/get-grouped-stories";

export const useMyStoriesGrouped = (
  groupBy: GroupedStoryParams["groupBy"] = "status",
  options?: Partial<GroupedStoryParams>,
  { accountScoped = false }: { accountScoped?: boolean } = {},
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  const params: GroupedStoryParams = {
    groupBy,
    ...options,
  };

  const baseKey = storyKeys.mineGrouped(workspaceSlug, params);
  const queryKey = accountScoped
    ? [...baseKey, "account", session?.user.id ?? ""]
    : baseKey;

  return useQuery({
    queryKey,
    queryFn: () =>
      getGroupedStories({ session: session!, workspaceSlug }, params),
    // Saved views require a fresh response whenever a filter key becomes active.
    staleTime: accountScoped ? 0 : 1000 * 60 * 2,
    enabled: !accountScoped || Boolean(session?.user.id),
    refetchOnMount: accountScoped ? "always" : true,
  });
};
