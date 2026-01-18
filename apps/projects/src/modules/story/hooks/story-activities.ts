import { useInfiniteQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStoryActivities } from "../queries/get-activities";

export const useStoryActivitiesInfinite = (id: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useInfiniteQuery({
    queryKey: storyKeys.activitiesInfinite(workspaceSlug, id),
    queryFn: ({ pageParam }) =>
      getStoryActivities(id, { session: session!, workspaceSlug }, pageParam),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
