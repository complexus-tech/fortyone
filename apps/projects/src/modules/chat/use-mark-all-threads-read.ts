import type { QueryClient } from "@tanstack/react-query";
import {
  useIsMutating,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { markAllThreadsRead } from "./api";
import { chatKeys } from "./constants";
import { reconcileChatQueries } from "./realtime";

const pendingRequests = new WeakMap<QueryClient, Map<string, Promise<void>>>();

export const useMarkAllThreadsRead = () => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const userId = session?.user.id ?? "";
  const queryClient = useQueryClient();
  const mutationKey = [
    ...chatKeys.all(workspaceSlug),
    "mark-all-threads-read",
    userId,
  ];
  const pendingCount = useIsMutating({ mutationKey, exact: true });
  const mutation = useMutation({
    mutationKey,
    mutationFn: markAllThreadsRead,
    // A retry takes a new server snapshot and could consume newly arrived replies.
    retry: false,
    onSuccess: async (_data, scope) => {
      // Refetch instead of zeroing counts: newer replies and unread timeline
      // messages must survive. Use the workspace captured when the action began.
      await reconcileChatQueries(queryClient, scope.workspaceSlug);
    },
  });

  return {
    isPending: pendingCount > 0,
    mutateAsync: () => {
      if (!userId || !workspaceSlug) {
        return Promise.reject(new Error("Sign in to mark threads as read."));
      }
      const scope = JSON.stringify(mutationKey);
      const requests =
        pendingRequests.get(queryClient) ?? new Map<string, Promise<void>>();
      const pending = requests.get(scope);
      if (pending) return pending;

      // The conversation header remounts when switching rooms. Share both the
      // pending indicator and request so another click cannot take a new snapshot.
      pendingRequests.set(queryClient, requests);
      const request = mutation.mutateAsync({ workspaceSlug }).finally(() => {
        requests.delete(scope);
        if (requests.size === 0) pendingRequests.delete(queryClient);
      });
      requests.set(scope, request);
      return request;
    },
  };
};
