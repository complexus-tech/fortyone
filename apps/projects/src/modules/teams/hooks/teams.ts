import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { Team } from "../types";
import { getTeams, getTeamsPage } from "../queries/get-teams";
import { getPublicTeams, getPublicTeamsPage } from "../queries/get-public-teams";

export const TEAM_MENU_PAGE_SIZE = 15;

export const useTeams = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery<Team[]>({
    queryKey: teamKeys.lists(workspaceSlug),
    queryFn: () => getTeams({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const usePublicTeams = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery<Team[]>({
    queryKey: teamKeys.public(workspaceSlug),
    queryFn: () => getPublicTeams({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useTeamsInfinite = (
  search = "",
  pageSize = TEAM_MENU_PAGE_SIZE,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useInfiniteQuery({
    queryKey: [
      ...teamKeys.lists(workspaceSlug),
      "infinite",
      normalizedSearch,
      pageSize,
    ] as const,
    queryFn: ({ pageParam }) =>
      getTeamsPage(
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

export const usePublicTeamsInfinite = (
  search = "",
  pageSize = TEAM_MENU_PAGE_SIZE,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useInfiniteQuery({
    queryKey: [
      ...teamKeys.public(workspaceSlug),
      "infinite",
      normalizedSearch,
      pageSize,
    ] as const,
    queryFn: ({ pageParam }) =>
      getPublicTeamsPage(
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
