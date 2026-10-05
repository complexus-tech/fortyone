"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { useInfiniteQuery, useQueries } from "@tanstack/react-query";
import { z } from "zod";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { teamKeys } from "@/constants/keys";
import { getTeamsPage } from "@/modules/teams/public/queries";
import { ViewsPresenceProvider } from "@/shared/views/presence-context";
import { presetKey } from "./hooks";
import { hasSavedViews } from "./view-presence";

/** Sidebar composition shares scoped presence without depending on feature APIs. */
export const SavedViewsPresenceProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const userId = session?.user.id ?? "";
  const enabled = Boolean(userId && workspaceSlug);
  const directory = useInfiniteQuery({
    queryKey: [...teamKeys.lists(workspaceSlug), "views", userId],
    queryFn: ({ pageParam }) =>
      getTeamsPage({ session, workspaceSlug }, "", pageParam, 15),
    initialPageParam: 1,
    getNextPageParam: (page, _pages, _parameter, parameters) =>
      page.pagination.hasMore && !parameters.includes(page.pagination.nextPage)
        ? page.pagination.nextPage
        : undefined,
    enabled,
    staleTime: 0,
    refetchOnMount: "always",
    retry: 1,
  });
  const teams =
    enabled && directory.isSuccess && directory.isFetchedAfterMount
      ? Array.from(
          new Set(
            directory.data.pages.flatMap((page) =>
              page.teams.map((team) => team.id),
            ),
          ),
        ).filter((teamId) => z.uuid().safeParse(teamId).success)
      : [];
  const queries = useQueries({
    queries: teams.map((teamId) => ({
      queryKey: [
        ...presetKey(workspaceSlug, userId, teamId, "view"),
        "presence",
      ],
      queryFn: ({ signal }: { signal: AbortSignal }) =>
        hasSavedViews(teamId, { session, workspaceSlug }, signal),
      staleTime: 0,
      refetchOnMount: "always" as const,
      retry: 1,
    })),
  });
  const batchSettled = queries.every(
    (query) => query.isError || query.isFetchedAfterMount,
  );
  const {
    fetchNextPage,
    hasNextPage,
    isFetching,
    isFetchedAfterMount,
    isError,
  } = directory;
  useEffect(() => {
    if (
      enabled &&
      isFetchedAfterMount &&
      !isError &&
      !isFetching &&
      hasNextPage &&
      batchSettled
    )
      void fetchNextPage({ cancelRefetch: false });
  }, [
    enabled,
    isFetchedAfterMount,
    isError,
    isFetching,
    hasNextPage,
    batchSettled,
    fetchNextPage,
  ]);
  const viewTeams = new Set(
    teams.filter((_teamId, index) => {
      const query = queries[index];
      return query.isSuccess && query.isFetchedAfterMount && query.data;
    }),
  );
  return (
    <ViewsPresenceProvider
      value={{
        hasViews: viewTeams.size > 0,
        hasTeamViews: (teamId) => viewTeams.has(teamId),
      }}
    >
      {children}
    </ViewsPresenceProvider>
  );
};
