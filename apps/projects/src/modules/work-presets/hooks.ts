"use client";

import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { archivePreset, createPreset, listPresets, renamePreset } from "./api";
import type { PresetInput, PresetKind } from "./types";

export const presetKey = (
  workspace: string,
  userId: string,
  teamId: string,
  kind: PresetKind,
) => ["work-presets", workspace, userId, teamId, kind] as const;

export const useWorkPresets = (
  teamId: string,
  kind: PresetKind,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useInfiniteQuery({
    queryKey: presetKey(workspaceSlug, session?.user.id ?? "", teamId, kind),
    queryFn: ({ pageParam, signal }) =>
      listPresets(teamId, kind, pageParam, { session, workspaceSlug }, signal),
    initialPageParam: "",
    getNextPageParam: (page, _pages, _parameter, parameters) =>
      page.nextCursor && !parameters.includes(page.nextCursor)
        ? page.nextCursor
        : undefined,
    enabled: Boolean(session && teamId && enabled),
    staleTime: 60_000,
  });
};

export const usePresetMutations = (teamId: string, kind: PresetKind) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const ctx = { session, workspaceSlug };
  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: presetKey(workspaceSlug, session?.user.id ?? "", teamId, kind),
    });
  return {
    create: useMutation({
      mutationFn: (input: PresetInput) => createPreset(input, ctx),
      onSuccess: invalidate,
    }),
    rename: useMutation({
      mutationFn: ({ id, name }: { id: string; name: string }) =>
        renamePreset(id, name, ctx),
      onSuccess: invalidate,
    }),
    archive: useMutation({
      mutationFn: (id: string) => archivePreset(id, ctx),
      onSuccess: invalidate,
    }),
  };
};
