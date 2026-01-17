import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { notificationKeys } from "@/constants/keys";
import { readNotification } from "../actions/read";
import type { AppNotification } from "../types";

/**
 * This hook is used to read a notification.
 * It is optimistic by default, meaning it will update the notification as read even if the server request fails.
 * @param isOptimistic - Whether to use optimistic updates.
 * @returns A mutation object.
 */

export const useReadNotificationMutation = (isOptimistic = true) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (notificationId: string) => readNotification(notificationId, workspaceSlug),

    onMutate: async (notificationId) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: notificationKeys.all(workspaceSlug) });

      // Get the previous data
      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all(workspaceSlug),
      );

      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread(workspaceSlug),
      );

      if (previousUnreadCount && isOptimistic) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(workspaceSlug),
          previousUnreadCount - 1,
        );
      }

      // Optimistically update the notifications
      if (previousNotifications && isOptimistic) {
        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all(workspaceSlug),
          previousNotifications.map((notification) =>
            notification.id === notificationId
              ? { ...notification, readAt: new Date().toISOString() }
              : notification,
          ),
        );
      }

      return { previousNotifications, previousUnreadCount };
    },

    onError: (error, notificationId, context) => {
      // Rollback on error
      if (context?.previousNotifications) {
        queryClient.setQueryData(
          notificationKeys.all(workspaceSlug),
          context.previousNotifications,
        );
      }

      if (context?.previousUnreadCount) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(workspaceSlug),
          context.previousUnreadCount,
        );
      }
      toast.error("Failed to mark notification as read", {
        description: error.message || "Please try again",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(notificationId);
          },
        },
      });
    },

    onSettled: () => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: notificationKeys.all(workspaceSlug) });
      queryClient.invalidateQueries({ queryKey: notificationKeys.unread(workspaceSlug) });
    },
  });

  return mutation;
};
