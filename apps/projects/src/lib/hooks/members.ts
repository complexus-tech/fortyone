import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import {
  getMembers,
  getMembersPage,
} from "@/lib/queries/members/get-members";
import { memberKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const MEMBER_MENU_PAGE_SIZE = 15;

export const useMembers = (search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useQuery({
    queryKey: [...memberKeys.lists(workspaceSlug), normalizedSearch],
    queryFn: () =>
      getMembers({ session: session!, workspaceSlug }, normalizedSearch),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useMembersInfinite = (
  search = "",
  pageSize = MEMBER_MENU_PAGE_SIZE,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useInfiniteQuery({
    queryKey: [
      ...memberKeys.lists(workspaceSlug),
      "infinite",
      normalizedSearch,
      pageSize,
    ] as const,
    queryFn: ({ pageParam }) =>
      getMembersPage(
        { session: session!, workspaceSlug },
        normalizedSearch,
        pageParam,
        pageSize,
      ),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    enabled: Boolean(session && enabled),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
