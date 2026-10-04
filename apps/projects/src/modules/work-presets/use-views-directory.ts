"use client";

import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { z } from "zod";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { teamKeys } from "@/constants/keys";
import { getTeamDetails, getTeamsPage } from "@/modules/teams/public/queries";
import { useViewCatalog } from "./use-view-catalog";
import type { ViewCatalogTeam } from "./use-view-catalog";

/** A scoped directory with fresh membership checks and recoverable cursor chains. */
export const useViewsDirectory = (
  teamId: string | null = null,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const userId = session?.user.id ?? "";
  const ctx = { session, workspaceSlug };
  const authenticated = Boolean(userId && workspaceSlug && enabled);
  const validTeamId = Boolean(teamId && z.uuid().safeParse(teamId).success);
  const scopedTeam = useQuery({
    queryKey: [...teamKeys.lists(workspaceSlug), "views", userId, teamId],
    queryFn: ({ signal }) => getTeamDetails(teamId!, ctx, signal),
    enabled: authenticated && validTeamId,
    staleTime: 0,
    refetchOnMount: "always",
    retry: 1,
  });
  const directory = useInfiniteQuery({
    queryKey: [...teamKeys.lists(workspaceSlug), "views", userId],
    queryFn: ({ pageParam }) => getTeamsPage(ctx, "", pageParam, 15),
    initialPageParam: 1,
    getNextPageParam: (page, _pages, _parameter, parameters) =>
      page.pagination.hasMore && !parameters.includes(page.pagination.nextPage)
        ? page.pagination.nextPage
        : undefined,
    enabled: authenticated && !teamId,
    staleTime: 0,
    refetchOnMount: "always",
    retry: 1,
  });
  const teamsError = teamId
    ? !validTeamId || scopedTeam.isError
    : directory.isError;
  const teamsFresh = teamId
    ? scopedTeam.isFetchedAfterMount
    : directory.isFetchedAfterMount;
  let teams: ViewCatalogTeam[] = [];
  if (authenticated && !teamsError && teamsFresh) {
    if (teamId) {
      if (scopedTeam.data) teams = [scopedTeam.data];
    } else teams = directory.data?.pages.flatMap((page) => page.teams) ?? [];
  }
  const { feeds } = useViewCatalog(teams);
  const readyFeeds = feeds.filter(
    (feed) => feed.isFetchedAfterMount && !feed.isError,
  );
  const views = readyFeeds.flatMap((feed) =>
    feed.views.map((view) => ({ view, team: feed.team })),
  );
  const pending =
    !teamsError &&
    (!authenticated ||
      !teamsFresh ||
      feeds.some((feed) => !feed.isFetchedAfterMount && !feed.isError));
  const failedFeeds = feeds.filter((feed) => feed.isError);
  const hasMoreTeams = !teamId && Boolean(directory.hasNextPage);
  const hasMoreViews = readyFeeds.some((feed) => feed.hasMore);
  const loadingMoreTeams = directory.isFetchingNextPage;
  const complete =
    !pending &&
    !teamsError &&
    !failedFeeds.length &&
    !hasMoreTeams &&
    !hasMoreViews &&
    !feeds.some((feed) => feed.isFetching);
  return {
    teams,
    views,
    feeds,
    pending,
    teamsError,
    failedFeeds,
    hasMoreTeams,
    hasMoreViews,
    loadingMoreTeams,
    complete,
    loadMoreTeams: () => {
      if (directory.hasNextPage && !directory.isFetchingNextPage)
        void directory.fetchNextPage({ cancelRefetch: false });
    },
    retryTeams: () =>
      void (teamId ? scopedTeam.refetch() : directory.refetch()),
  };
};
