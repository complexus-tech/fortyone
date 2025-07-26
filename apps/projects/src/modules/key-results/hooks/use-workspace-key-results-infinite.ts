import { useInfiniteQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getWorkspaceKeyResults } from "../queries/get-workspace-key-results";
import type { KeyResultFilters, KeyResultListResponse } from "../types";

const keyResultKeys = {
  all: ["key-results"] as const,
  lists: () => [...keyResultKeys.all, "list"] as const,
  list: (filters?: KeyResultFilters) =>
    [...keyResultKeys.lists(), filters] as const,
};

export const useWorkspaceKeyResultsInfinite = (filters?: KeyResultFilters) => {
  const { data: session } = useSession();

  return useInfiniteQuery({
    queryKey: keyResultKeys.list(filters),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    queryFn: ({ pageParam }) =>
      getWorkspaceKeyResults(session!, {
        ...filters,
        page: pageParam,
      }),
    getNextPageParam: (lastPage: KeyResultListResponse) =>
      lastPage.hasMore ? lastPage.page + 1 : undefined,
    initialPageParam: 1,
  });
};
