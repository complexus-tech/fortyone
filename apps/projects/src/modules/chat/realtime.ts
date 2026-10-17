import type { QueryClient } from "@tanstack/react-query";
import { chatKeys } from "./constants";

type PendingRefresh = {
  conversations: Set<string>;
  all: boolean;
  promise: Promise<void>;
};
const pendingRefreshes = new WeakMap<
  QueryClient,
  Map<string, PendingRefresh>
>();

/** Coalesce SSE hints and reconcile using authorized HTTP reads. */
export const reconcileChatQueries = (
  queryClient: QueryClient,
  workspaceSlug: string,
  conversationId?: string,
): Promise<void> => {
  const workspace = workspaceSlug.toLowerCase();
  const refreshes =
    pendingRefreshes.get(queryClient) ?? new Map<string, PendingRefresh>();
  pendingRefreshes.set(queryClient, refreshes);
  const pending = refreshes.get(workspace);
  if (pending) {
    if (conversationId) pending.conversations.add(conversationId);
    else pending.all = true;
    return pending.promise;
  }
  const conversations = new Set(conversationId ? [conversationId] : []);
  const refresh: PendingRefresh = {
    conversations,
    all: !conversationId,
    promise: Promise.resolve(),
  };
  refresh.promise = new Promise<void>((resolve, reject) => {
    setTimeout(() => {
      refreshes.delete(workspace);
      void queryClient
        .invalidateQueries({
          queryKey: chatKeys.all(workspace),
          predicate: ({ queryKey }: { queryKey: readonly unknown[] }) =>
            refresh.all ||
            queryKey[2] === "unread" ||
            queryKey[2] === "unread-threads" ||
            queryKey[2] === "conversations" ||
            queryKey[2] === "message" ||
            (typeof queryKey[3] === "string" && conversations.has(queryKey[3])),
        })
        .then(() => {
          resolve();
        }, reject);
    }, 150);
  });
  refreshes.set(workspace, refresh);
  return refresh.promise;
};
