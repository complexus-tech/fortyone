import { useInfiniteQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { keyResultKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useWorkspacePath } from "@/hooks";
import { getWorkspaceKeyResults } from "../queries/get-workspace-key-results";
import type { KeyResultFilters, KeyResultListResponse } from "../types";

export const useWorkspaceKeyResultsInfinite = (filters?: KeyResultFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useInfiniteQuery({
    queryKey: keyResultKeys.list(workspaceSlug, filters),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    queryFn: ({ pageParam }) =>
      getWorkspaceKeyResults(
        { session: session!, workspaceSlug },
        {
          ...filters,
          page: pageParam,
        },
      ),
    getNextPageParam: (lastPage: KeyResultListResponse) =>
      lastPage.hasMore ? lastPage.page + 1 : undefined,
    initialPageParam: 1,
    enabled: Boolean(session),
  });
};
