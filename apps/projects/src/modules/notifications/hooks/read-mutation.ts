import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { notificationKeys } from "@/constants/keys";
import { readNotification } from "../actions/read";
import type { AppNotification } from "../types";

export const useReadNotificationMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (notificationId: string) => readNotification(notificationId),

    onMutate: async (notificationId) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: notificationKeys.all });

      // Get the previous data
      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all,
      );

      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread(),
      );

      if (previousUnreadCount) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(),
          previousUnreadCount - 1,
        );
      }

      // Optimistically update the notifications
      if (previousNotifications) {
        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all,
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
          notificationKeys.all,
          context.previousNotifications,
        );
      }

      if (context?.previousUnreadCount) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(),
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
      queryClient.invalidateQueries({ queryKey: notificationKeys.all });
      queryClient.invalidateQueries({ queryKey: notificationKeys.unread() });
    },
  });

  return mutation;
};
