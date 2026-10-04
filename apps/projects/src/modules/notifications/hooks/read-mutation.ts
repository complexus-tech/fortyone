import type { QueryKey } from "@tanstack/react-query";
import type { ReadNotificationInput } from "lib/src/notification-read";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { notificationKeys } from "@/constants/keys";
import { readNotification } from "../actions/read";
import {
  isNotificationReadCache,
  markNotificationReadInCache,
} from "../utils/read-cache";
import type { NotificationReadCache } from "../utils/read-cache";

type CacheSnapshot = {
  key: QueryKey;
  before: NotificationReadCache;
  after: unknown;
};

export const useReadNotificationMutation = (isOptimistic = true) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const allKey = notificationKeys.all(workspaceSlug);
  const unreadKey = notificationKeys.unread(workspaceSlug);

  const mutation = useMutation({
    mutationFn: async ({ id, observedCreatedAt }: ReadNotificationInput) => {
      const response = await readNotification(
        id,
        workspaceSlug,
        observedCreatedAt,
      );
      if (response?.error) throw new Error(response.error.message);
    },

    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: allKey });
      if (!isOptimistic) return undefined;
      const snapshots: CacheSnapshot[] = [];
      let changed = false;
      const readAt = new Date().toISOString();
      for (const [key, before] of queryClient.getQueriesData({
        queryKey: allKey,
      })) {
        if (!isNotificationReadCache(before)) continue;
        const next = markNotificationReadInCache(before, input, readAt);
        if (!next.changed) continue;
        changed = true;
        queryClient.setQueryData(key, next.data);
        snapshots.push({ key, before, after: queryClient.getQueryData(key) });
      }
      const beforeUnread = queryClient.getQueryData<number>(unreadKey);
      const afterUnread =
        changed && beforeUnread !== undefined
          ? Math.max(0, beforeUnread - 1)
          : beforeUnread;
      if (afterUnread !== beforeUnread) {
        queryClient.setQueryData(unreadKey, afterUnread);
      }
      return { snapshots, beforeUnread, afterUnread, unreadKey, allKey };
    },

    onError: (error, input, context) => {
      for (const snapshot of context?.snapshots ?? []) {
        // Preserve refreshes and mutations that arrived after this displayed version.
        if (queryClient.getQueryData(snapshot.key) === snapshot.after) {
          queryClient.setQueryData(snapshot.key, snapshot.before);
        }
      }
      if (
        context &&
        queryClient.getQueryData(context.unreadKey) === context.afterUnread
      ) {
        queryClient.setQueryData(context.unreadKey, context.beforeUnread);
      }
      toast.error("Failed to mark notification as read", {
        description: error.message || "Please try again",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(input);
          },
        },
      });
    },

    onSettled: (_data, _error, _input, context) =>
      queryClient.invalidateQueries({ queryKey: context?.allKey ?? allKey }),
  });

  return mutation;
};
