import { useInfiniteQuery, useIsMutating } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { commentMutationKey } from "@/shared/story/comment-cache";
import { getStoryComments } from "../queries/get-comments";

export const useStoryCommentsInfinite = (id: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const pendingChanges = useIsMutating({
    mutationKey: commentMutationKey(workspaceSlug),
    predicate: (mutation) =>
      (mutation.state.variables as { storyId?: string } | undefined)
        ?.storyId === id,
  });

  return useInfiniteQuery({
    queryKey: storyKeys.commentsInfinite(workspaceSlug, id),
    queryFn: ({ pageParam }) =>
      getStoryComments(id, { session: session!, workspaceSlug }, pageParam),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    refetchInterval: pendingChanges ? false : 30_000,
  });
};
