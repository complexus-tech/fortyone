import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { notificationKeys } from "@/constants/keys";
import { readAllNotifications } from "../actions/read-all";
import type { AppNotification } from "../types";

export const useReadAllNotificationsMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: () => readAllNotifications(workspaceSlug),

    onMutate: async () => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: notificationKeys.all(workspaceSlug) });

      // Get the previous data
      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all(workspaceSlug),
      );

      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread(workspaceSlug),
      );

      // Optimistically update all notifications as read
      if (previousNotifications) {
        const now = new Date().toISOString();
        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all(workspaceSlug),
          previousNotifications.map((notification) => ({
            ...notification,
            readAt: notification.readAt || now,
          })),
        );
      }

      // Also update the unread count to 0
      queryClient.setQueryData(notificationKeys.unread(workspaceSlug), 0);

      return { previousNotifications, previousUnreadCount };
    },

    onError: (error, _, context) => {
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
      toast.error("Failed to mark all notifications as read", {
        description: error.message || "Please try again",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate();
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
