"use client";

import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { presetKey } from "./hooks";
import { resolveSavedViews } from "./resolve-views";

export const useSelectedView = (
  teamId: string,
  viewId: string | null,
  selection = 0,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: [
      ...presetKey(workspaceSlug, session?.user.id ?? "", teamId, "view"),
      "selected",
      viewId,
      selection,
    ],
    queryFn: async ({ signal }) => {
      const views = await resolveSavedViews(
        teamId,
        [viewId!],
        { session, workspaceSlug },
        signal,
      );
      return views[0] ?? null;
    },
    enabled: Boolean(session && teamId && viewId),
    staleTime: 0,
    refetchOnMount: "always",
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    retry: 1,
  });
};
