import { useEffect, useRef } from "react";
import {
  useIsMutating,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { storyKeys } from "@/shared/story/cache-keys";
import { invalidateCommentQueries } from "@/shared/story/comment-cache";
import {
  getStoryMayaRuns,
  updateCommentMayaRun,
} from "../queries/comment-maya";
import type { CommentMayaAction, CommentMayaRun } from "../types/comment-maya";
import { isMayaRunActive } from "../types/comment-maya";

export const commentMayaKey = (workspaceSlug: string, storyId: string) =>
  storyKeys.mayaRuns(workspaceSlug, storyId);

export const useStoryMayaRuns = (storyId: string, { poll = false } = {}) => {
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: commentMayaKey(workspaceSlug, storyId),
    queryFn: () => getStoryMayaRuns(storyId, { workspaceSlug }),
    staleTime: 10_000,
    refetchInterval: (query) => {
      if (!poll) return false;
      return query.state.data?.runs.some(isMayaRunActive) ? 2_000 : 30_000;
    },
    refetchIntervalInBackground: false,
    retry: false,
  });
};

/** Run from the feed once; composers share the query without repeating refreshes. */
export const useMayaRunRefresh = (
  storyId: string,
  runs: CommentMayaRun[] | undefined,
) => {
  const client = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const previous = useRef(new Map<string, string>());
  const pendingStoryChanges = useIsMutating({
    predicate: (mutation) =>
      (mutation.state.variables as { storyId?: string } | undefined)
        ?.storyId === storyId,
  });
  useEffect(() => {
    if (!runs || pendingStoryChanges) return;
    let changed = false;
    for (const run of runs) {
      const marker = `${run.status}:${run.version}:${run.replyCommentId ?? ""}`;
      if (
        (!isMayaRunActive(run) || run.replyCommentId) &&
        previous.current.get(run.id) !== marker
      )
        changed = true;
      previous.current.set(run.id, marker);
    }
    if (!changed) return;
    void invalidateCommentQueries(client, workspaceSlug, storyId);
    void client.invalidateQueries({
      queryKey: storyKeys.detail(workspaceSlug, storyId),
      exact: true,
    });
    void client.invalidateQueries({
      queryKey: storyKeys.activitiesInfinite(workspaceSlug, storyId),
      refetchType: "all",
    });
    void client.invalidateQueries({
      queryKey: storyKeys.all(workspaceSlug),
      predicate: (query) => query.queryKey.at(2) !== "detail",
    });
    void client.invalidateQueries({ queryKey: storyKeys.total(workspaceSlug) });
  }, [client, pendingStoryChanges, runs, storyId, workspaceSlug]);
};

export const useCommentMayaAction = (storyId: string) => {
  const client = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: ({
      run,
      action,
    }: {
      run: CommentMayaRun;
      action: CommentMayaAction;
    }) =>
      updateCommentMayaRun(storyId, run.id, action, run.version, {
        workspaceSlug,
      }),
    onError: (error) => {
      toast.error("Could not update Maya's request", {
        description: error.message,
      });
    },
    onSettled: () =>
      client.invalidateQueries({
        queryKey: commentMayaKey(workspaceSlug, storyId),
      }),
  });
};
