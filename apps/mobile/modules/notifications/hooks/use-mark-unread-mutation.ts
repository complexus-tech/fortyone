import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner-native";
import { notificationKeys } from "@/constants/keys";
import { markUnread } from "../actions/mark-unread";
import type { AppNotification } from "../types";

export const useMarkUnreadMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (notificationId: string) => markUnread(notificationId),

    onMutate: async (notificationId) => {
      await queryClient.cancelQueries({ queryKey: notificationKeys.all });

      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all
      );
      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread()
      );

      // Optimistically update the notification as unread
      if (previousNotifications) {
        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all,
          previousNotifications.map((notification) =>
            notification.id === notificationId
              ? { ...notification, readAt: null }
              : notification
          )
        );

        // Increment the unread count since we're marking a read notification as unread
        if (previousUnreadCount !== undefined) {
          queryClient.setQueryData<number>(
            notificationKeys.unread(),
            previousUnreadCount + 1
          );
        }
      }

      return { previousNotifications, previousUnreadCount };
    },

    onError: (error, notificationId, context) => {
      if (context?.previousNotifications) {
        queryClient.setQueryData(
          notificationKeys.all,
          context.previousNotifications
        );
      }

      if (context?.previousUnreadCount !== undefined) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(),
          context.previousUnreadCount
        );
      }

      toast.error("Failed to mark notification as unread", {
        description: error.message || "Please try again",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(notificationId),
        },
      });
    },

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.all });
      queryClient.invalidateQueries({ queryKey: notificationKeys.unread() });
      toast.success("Notification marked as unread");
    },
  });

  return mutation;
};
