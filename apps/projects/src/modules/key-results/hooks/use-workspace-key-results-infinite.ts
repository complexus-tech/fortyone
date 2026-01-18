import { useInfiniteQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getWorkspaceKeyResults } from "../queries/get-workspace-key-results";
import type { KeyResultFilters, KeyResultListResponse } from "../types";
import { useWorkspacePath } from "@/hooks";

const keyResultKeys = {
  all: (workspaceSlug: string) => ["key-results", workspaceSlug] as const,
  lists: (workspaceSlug: string) => [...keyResultKeys.all(workspaceSlug), "list"] as const,
  list: (workspaceSlug: string, filters?: KeyResultFilters) =>
    [...keyResultKeys.lists(workspaceSlug), filters] as const,
};

export const useWorkspaceKeyResultsInfinite = (filters?: KeyResultFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useInfiniteQuery({
    queryKey: keyResultKeys.list(workspaceSlug, filters),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    queryFn: ({ pageParam }) =>
      getWorkspaceKeyResults({ session: session!, workspaceSlug }, {
        ...filters,
        page: pageParam,
      }),
    getNextPageParam: (lastPage: KeyResultListResponse) =>
      lastPage.hasMore ? lastPage.page + 1 : undefined,
    initialPageParam: 1,
  });
};
