import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import {
  getTeamSprints,
  getTeamSprintsPage,
} from "../queries/get-team-sprints";

export const SPRINT_MENU_PAGE_SIZE = 15;

export const useTeamSprints = (teamId: string, search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();
  return useQuery({
    queryKey: [...sprintKeys.team(workspaceSlug, teamId), normalizedSearch],
    queryFn: () =>
      getTeamSprints(
        teamId,
        { session: session!, workspaceSlug },
        normalizedSearch,
      ),
    enabled: Boolean(teamId && session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};

export const useTeamSprintsInfinite = (
  teamId: string,
  search = "",
  pageSize = SPRINT_MENU_PAGE_SIZE,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useInfiniteQuery({
    queryKey: [
      ...sprintKeys.team(workspaceSlug, teamId),
      "infinite",
      normalizedSearch,
      pageSize,
    ] as const,
    queryFn: ({ pageParam }) =>
      getTeamSprintsPage(
        teamId,
        { session: session!, workspaceSlug },
        normalizedSearch,
        pageParam,
        pageSize,
      ),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    enabled: Boolean(teamId && session && enabled),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
