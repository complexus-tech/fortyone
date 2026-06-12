import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import {
  getTeamMembers,
  getTeamMembersPage,
} from "@/lib/queries/members/get-members";
import { memberKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { MEMBER_MENU_PAGE_SIZE } from "./members";

export const useTeamMembers = (teamId?: string, search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useQuery({
    queryKey: [
      ...memberKeys.team(workspaceSlug, teamId ?? ""),
      normalizedSearch,
    ],
    queryFn: () =>
      getTeamMembers(
        teamId!,
        { session: session!, workspaceSlug },
        normalizedSearch,
      ),
    enabled: Boolean(teamId && session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useTeamMembersInfinite = (
  teamId?: string,
  search = "",
  pageSize = MEMBER_MENU_PAGE_SIZE,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useInfiniteQuery({
    queryKey: [
      ...memberKeys.team(workspaceSlug, teamId ?? ""),
      "infinite",
      normalizedSearch,
      pageSize,
    ] as const,
    queryFn: ({ pageParam }) =>
      getTeamMembersPage(
        teamId ?? "",
        { session: session!, workspaceSlug },
        normalizedSearch,
        pageParam,
        pageSize,
      ),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    enabled: Boolean(teamId && session && enabled),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
