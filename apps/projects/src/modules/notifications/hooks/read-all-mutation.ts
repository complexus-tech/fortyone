import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { notificationKeys } from "@/constants/keys";
import { readAllNotifications } from "../actions/read-all";
import type { AppNotification } from "../types";

export const useReadAllNotificationsMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: readAllNotifications,

    onMutate: async () => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: notificationKeys.all });

      // Get the previous data
      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all,
      );

      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread(),
      );

      // Optimistically update all notifications as read
      if (previousNotifications) {
        const now = new Date().toISOString();
        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all,
          previousNotifications.map((notification) => ({
            ...notification,
            readAt: notification.readAt || now,
          })),
        );
      }

      // Also update the unread count to 0
      queryClient.setQueryData(notificationKeys.unread(), 0);

      return { previousNotifications, previousUnreadCount };
    },

    onError: (error, _, context) => {
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
      queryClient.invalidateQueries({ queryKey: notificationKeys.all });
      queryClient.invalidateQueries({ queryKey: notificationKeys.unread() });
    },
  });

  return mutation;
};
