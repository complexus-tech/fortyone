"use client";

import { useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { get, post } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import { documentKeys } from "@/shared/documents/keys";
import type { DocumentCommentThread } from "./types";

const REPLY_REFRESH_INTERVAL = 2_000;
const MAX_REPLY_REFRESHES = 5;

/** A published reply can precede a successful comments fetch. Retry that handoff
 * briefly, without making a missing/inaccessible comment poll indefinitely. */
function retryPublishedReplies(
  client: QueryClient,
  key: ReturnType<typeof documentKeys.comments>,
  replyIds: string[],
  attempts: Map<string, number>,
) {
  let disposed = false;
  const isDisposed = () => disposed;
  let timer: ReturnType<typeof setTimeout> | undefined;
  const missing = () => {
    const comments = client.getQueryData<DocumentCommentThread[]>(key) ?? [];
    const present = new Set(
      comments.flatMap((thread) =>
        thread.comments.map((comment) => comment.id),
      ),
    );
    return replyIds.filter(
      (id) => !present.has(id) && (attempts.get(id) ?? 0) < MAX_REPLY_REFRESHES,
    );
  };
  const refresh = async () => {
    const pending = missing();
    if (isDisposed() || !pending.length) return;
    for (const id of pending) attempts.set(id, (attempts.get(id) ?? 0) + 1);
    await client.invalidateQueries({ queryKey: key, exact: true });
    if (!isDisposed() && missing().length)
      timer = setTimeout(() => {
        void refresh();
      }, REPLY_REFRESH_INTERVAL);
  };
  if (missing().length)
    timer = setTimeout(() => {
      void refresh();
    }, REPLY_REFRESH_INTERVAL);
  return () => {
    disposed = true;
    clearTimeout(timer);
  };
}

export type DocumentMayaRun = {
  id: string;
  sourceId: string;
  requesterId: string;
  status:
    | "queued"
    | "running"
    | "needs_input"
    | "awaiting_approval"
    | "succeeded"
    | "failed"
    | "cancelled";
  replyCommentId?: string | null;
  summary: string;
  proposalReview?: string;
  updatedAt: string;
  version: number;
};
export type DocumentMayaRuns = { enabled: boolean; runs: DocumentMayaRun[] };
export const documentMayaKey = (workspaceSlug: string, documentId: string) =>
  [...documentKeys.detail(workspaceSlug, documentId), "maya-runs"] as const;
export const isDocumentMayaActive = (run: DocumentMayaRun) =>
  run.status === "queued" || run.status === "running";

/** The panel owns polling; composers and rows share this cache. */
export function useDocumentMayaRuns(documentId: string, poll = false) {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  return useQuery({
    queryKey: documentMayaKey(workspaceSlug, documentId),
    queryFn: async () => {
      const response = await get<ApiResponse<DocumentMayaRuns>>(
        `documents/${documentId}/maya-runs`,
        { workspaceSlug, session: session! },
      );
      if (!response.data)
        throw new Error("Could not check for Maya’s response");
      return response.data;
    },
    enabled: Boolean(session && documentId),
    staleTime: 5_000,
    refetchInterval: (query) => {
      if (!poll || query.state.data?.enabled === false) return false;
      return query.state.data?.runs.some(isDocumentMayaActive) ? 2_000 : 30_000;
    },
    refetchIntervalInBackground: false,
    retry: false,
  });
}

export function useDocumentMayaRefresh(
  documentId: string,
  runs?: DocumentMayaRun[],
) {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const seen = useRef(new Map<string, string>());
  const replyRefreshes = useRef(new Map<string, number>());
  useEffect(() => {
    if (!runs) return;
    let changed = false;
    for (const run of runs) {
      const marker = `${run.status}:${run.version}:${run.replyCommentId ?? ""}`;
      if (
        (!isDocumentMayaActive(run) || run.replyCommentId) &&
        seen.current.get(run.id) !== marker
      )
        changed = true;
      seen.current.set(run.id, marker);
    }
    const commentsKey = documentKeys.comments(workspaceSlug, documentId);
    if (changed) {
      void queryClient.invalidateQueries({
        queryKey: commentsKey,
        exact: true,
      });
      void queryClient.invalidateQueries({
        queryKey: documentKeys.detail(workspaceSlug, documentId),
        exact: true,
      });
      void queryClient.invalidateQueries({
        queryKey: documentKeys.lists(workspaceSlug),
      });
    }
    return retryPublishedReplies(
      queryClient,
      commentsKey,
      runs.flatMap((run) => (run.replyCommentId ? [run.replyCommentId] : [])),
      replyRefreshes.current,
    );
  }, [documentId, queryClient, runs, workspaceSlug]);
}

export function useDocumentMayaAction(documentId: string) {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  return useMutation({
    mutationFn: async ({
      run,
      action,
    }: {
      run: DocumentMayaRun;
      action: "retry" | "cancel" | "approve";
    }) => {
      const response = await post<
        { version: number },
        ApiResponse<DocumentMayaRun>
      >(
        `documents/${documentId}/maya-runs/${run.id}/${action}`,
        { version: run.version },
        { workspaceSlug, session: session! },
      );
      if (response.error) throw new Error(response.error.message);
      return response.data;
    },
    onError: (error) =>
      toast.error("Could not update Maya’s request", {
        description: error.message,
      }),
    onSettled: () =>
      queryClient.invalidateQueries({
        queryKey: documentMayaKey(workspaceSlug, documentId),
      }),
  });
}
