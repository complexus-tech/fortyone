"use client";

import { useState } from "react";
import { useQueries, useQueryClient } from "@tanstack/react-query";
import type { WorkspaceCtx } from "@/lib/http";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { listPresets } from "./api";
import { presetKey } from "./hooks";
import type { PresetPage } from "./types";
import type { SavedView } from "./resolve-views";

export type ViewCatalogTeam = { id: string; name: string };

/** Rebuild the requested page chain so refreshes never reuse expired or shifted cursors. */
const listViewPages = async (
  teamId: string,
  requestedPages: number,
  ctx: WorkspaceCtx,
  signal: AbortSignal,
) => {
  const pages: PresetPage[] = [];
  const seen = new Set<string>();
  let cursor = "";
  for (let page = 0; page < requestedPages; page++) {
    seen.add(cursor);
    // eslint-disable-next-line no-await-in-loop -- Each cursor comes from the preceding response.
    const result = await listPresets(teamId, "view", cursor, ctx, signal);
    pages.push(result);
    if (!result.nextCursor) break;
    if (seen.has(result.nextCursor))
      throw new Error("Views could not be loaded. Please try again.");
    cursor = result.nextCursor;
  }
  return pages;
};

export const useViewCatalog = (teams: ViewCatalogTeam[]) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const [pageCounts, setPageCounts] = useState<Record<string, number>>({});
  const queryKey = (teamId: string, count: number) => [
    ...presetKey(workspaceSlug, session?.user.id ?? "", teamId, "view"),
    "catalog",
    count,
  ];
  const queries = useQueries({
    queries: teams.map((team) => {
      const count = pageCounts[team.id] ?? 1;
      return {
        queryKey: queryKey(team.id, count),
        queryFn: ({ signal }: { signal: AbortSignal }) =>
          listViewPages(team.id, count, { session, workspaceSlug }, signal),
        placeholderData: () =>
          queryClient.getQueryData<PresetPage[]>(queryKey(team.id, count - 1)),
        enabled: Boolean(session),
        staleTime: 60_000,
        refetchOnMount: "always" as const,
      };
    }),
  });
  const feeds = teams.map((team, index) => {
    const query = queries[index];
    const pages = query.isError ? [] : query.data ?? [];
    const seen = new Set<string>();
    const views = pages
      .flatMap((page) => page.items)
      .filter((preset): preset is SavedView => {
        if (
          preset.kind !== "view" ||
          preset.teamId !== team.id ||
          seen.has(preset.id)
        )
          return false;
        seen.add(preset.id);
        return true;
      });
    const nextCursor = pages.at(-1)?.nextCursor;
    const count = pageCounts[team.id] ?? 1;
    return {
      team,
      views,
      isPending: query.isPending,
      isFetchedAfterMount: query.isFetchedAfterMount,
      isFetching: query.isFetching,
      isError: query.isError,
      retry: () => query.refetch(),
      hasMore: Boolean(nextCursor),
      loadMore: () => {
        if (!nextCursor || query.isFetching) return;
        setPageCounts((current) =>
          (current[team.id] ?? 1) === count
            ? { ...current, [team.id]: count + 1 }
            : current,
        );
      },
    };
  });
  return { feeds, userId: session?.user.id };
};
