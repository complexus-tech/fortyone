import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getObjectives } from "../queries/get-objectives";
import { objectiveKeys } from "../constants";
import {
  getTeamObjectives,
  getTeamObjectivesPage,
} from "../queries/get-team-objectives";

export const OBJECTIVE_MENU_PAGE_SIZE = 15;

export const useObjectives = (search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();
  return useQuery({
    queryKey: [...objectiveKeys.list(workspaceSlug), normalizedSearch],
    queryFn: () =>
      getObjectives({ session: session!, workspaceSlug }, normalizedSearch),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useTeamObjectives = (teamId: string, search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();
  return useQuery({
    queryKey: [...objectiveKeys.team(workspaceSlug, teamId), normalizedSearch],
    queryFn: () =>
      getTeamObjectives(
        teamId,
        { session: session!, workspaceSlug },
        normalizedSearch,
      ),
    enabled: Boolean(teamId && session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useTeamObjectivesInfinite = (
  teamId: string,
  search = "",
  pageSize = OBJECTIVE_MENU_PAGE_SIZE,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useInfiniteQuery({
    queryKey: [
      ...objectiveKeys.team(workspaceSlug, teamId),
      "infinite",
      normalizedSearch,
      pageSize,
    ] as const,
    queryFn: ({ pageParam }) =>
      getTeamObjectivesPage(
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
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
