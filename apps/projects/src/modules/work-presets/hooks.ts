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

const presetKey = (workspace: string, teamId: string, kind: PresetKind) =>
  ["work-presets", workspace, teamId, kind] as const;

export const useWorkPresets = (
  teamId: string,
  kind: PresetKind,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useInfiniteQuery({
    queryKey: presetKey(workspaceSlug, teamId, kind),
    queryFn: ({ pageParam }) =>
      listPresets(teamId, kind, pageParam, { session, workspaceSlug }),
    initialPageParam: "",
    getNextPageParam: (page) => page.nextCursor || undefined,
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
      queryKey: presetKey(workspaceSlug, teamId, kind),
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
