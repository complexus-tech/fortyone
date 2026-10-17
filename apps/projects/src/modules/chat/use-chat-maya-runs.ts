"use client";

import { useEffect, useRef } from "react";
import {
  useMutation,
  useMutationState,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { get, post } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import { chatKeys } from "./constants";

export type ChatMayaRun = {
  id: string;
  sourceId: string;
  /** Destination thread; optional while older API replicas are being replaced. */
  threadRootId?: string;
  requesterId: string;
  status:
    | "queued"
    | "running"
    | "needs_input"
    | "awaiting_approval"
    | "succeeded"
    | "failed"
    | "cancelled";
  summary: string;
  proposalReview?: string;
  replyMessageId?: string | null;
  version: number;
  updatedAt: string;
  tasks?: {
    id: string;
    reference: string;
    title: string;
    teamId: string;
    priority: string;
    statusId?: string;
    assigneeId?: string;
  }[];
};
type ChatMayaRuns = { enabled: boolean; runs: ChatMayaRun[] };
type MayaAction = "approve" | "cancel" | "retry";
type MayaActionInput = { run: ChatMayaRun; action: MayaAction };
const active = (run: ChatMayaRun) =>
  run.status === "queued" || run.status === "running";

/** Only the conversation shell polls; virtualized rows share its query cache. */
export function useChatMayaRuns(
  conversationId: string,
  { poll = false }: { poll?: boolean } = {},
) {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const client = useQueryClient();
  const observed = useRef(new Map<string, string>());
  const query = useQuery({
    queryKey: chatKeys.mayaRuns(workspaceSlug, conversationId),
    queryFn: async () => {
      const response = await get<ApiResponse<ChatMayaRuns>>(
        `chat/conversations/${conversationId}/maya-runs`,
        { workspaceSlug, session: session! },
      );
      if (!response.data) throw new Error("Could not check Maya’s response");
      return response.data;
    },
    enabled: Boolean(session && conversationId),
    staleTime: 5_000,
    refetchInterval: (state) => {
      if (!poll || state.state.data?.enabled === false) return false;
      return state.state.data?.runs.some(active) ? 2_000 : 30_000;
    },
    refetchIntervalInBackground: false,
    retry: false,
  });
  useEffect(() => {
    if (!poll) return;
    let changed = false;
    for (const run of query.data?.runs ?? []) {
      const key = `${workspaceSlug}:${conversationId}:${run.id}`;
      const marker = `${run.version}:${run.status}:${run.replyMessageId ?? ""}`;
      if (run.replyMessageId && observed.current.get(key) !== marker)
        changed = true;
      observed.current.set(key, marker);
    }
    if (changed) {
      for (const kind of ["messages", "replies"])
        void client.invalidateQueries({
          queryKey: [...chatKeys.all(workspaceSlug), kind, conversationId],
        });
    }
  }, [client, conversationId, poll, query.data, workspaceSlug]);
  return query;
}

export function useChatMayaAction(conversationId: string, runId?: string) {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const client = useQueryClient();
  const queryKey = chatKeys.mayaRuns(workspaceSlug, conversationId);
  const mutationKey = [...queryKey, "action", session?.user.id, runId];
  const pending = useMutationState({
    filters: { mutationKey, exact: true, status: "pending" },
    select: (mutation) => (mutation.state.variables as MayaActionInput).action,
  });
  const mutation = useMutation({
    mutationKey,
    mutationFn: async ({ run, action }: MayaActionInput) => {
      const response = await post<
        { version: number },
        ApiResponse<ChatMayaRun>
      >(
        `chat/conversations/${conversationId}/maya-runs/${run.id}/${action}`,
        { version: run.version },
        { workspaceSlug, session: session! },
      );
      if (response.error) throw new Error(response.error.message);
      if (!response.data) throw new Error("Could not update Maya’s request");
      return response.data;
    },
    onMutate: () => client.cancelQueries({ queryKey, exact: true }),
    onSuccess: async (updated) => {
      await client.cancelQueries({ queryKey, exact: true }, { revert: false });
      client.setQueryData<ChatMayaRuns>(queryKey, (current) =>
        current
          ? {
              ...current,
              runs: current.runs.map((run) =>
                run.id === updated.id && run.version <= updated.version
                  ? updated
                  : run,
              ),
            }
          : current,
      );
    },
    onError: (error) =>
      toast.error("Could not update Maya’s request", {
        description: error.message,
      }),
    onSettled: () => client.invalidateQueries({ queryKey, exact: true }),
  });
  const pendingAction = pending.at(0);
  return {
    isPending: pendingAction !== undefined,
    pendingAction,
    mutate: (input: MayaActionInput) => {
      // Main-message and thread copies share one request. This synchronous
      // cache check also guards clicks before React has rendered the busy state.
      if (client.isMutating({ mutationKey, exact: true }) > 0) return;
      mutation.mutate(input);
    },
  };
}
