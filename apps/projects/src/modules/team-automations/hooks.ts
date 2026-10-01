"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import {
  archiveAutomation,
  createAutomation,
  listAutomations,
  listRuns,
  pauseAutomation,
} from "./api";
import type { AutomationInput } from "./types";

export const useAutomations = (teamId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const client = useQueryClient();
  const ctx = { session, workspaceSlug };
  const key = ["team-automations", workspaceSlug, teamId];
  const invalidate = () => client.invalidateQueries({ queryKey: key });
  const query = useQuery({
    queryKey: key,
    queryFn: () => listAutomations(teamId, ctx),
    enabled: Boolean(session && teamId),
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
  return {
    ...query,
    create: useMutation({
      mutationFn: (input: AutomationInput) => createAutomation(input, ctx),
      onSuccess: invalidate,
    }),
    pause: useMutation({
      mutationFn: ({ id, paused }: { id: string; paused: boolean }) =>
        pauseAutomation(id, paused, ctx),
      onSuccess: invalidate,
    }),
    archive: useMutation({
      mutationFn: (id: string) => archiveAutomation(id, ctx),
      onSuccess: invalidate,
    }),
  };
};
export const useAutomationRuns = (id: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: ["team-automation-runs", workspaceSlug, id],
    queryFn: () => listRuns(id, { session, workspaceSlug }),
    enabled: Boolean(session && id),
    staleTime: 10_000,
    refetchInterval: 10_000,
  });
};
