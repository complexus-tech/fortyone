import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Alert } from "react-native";
import { notificationKeys } from "@/constants/keys";
import { deleteNotification } from "../actions/delete";
import type { AppNotification } from "../types";

export const useDeleteMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (notificationId: string) => deleteNotification(notificationId),

    onMutate: async (notificationId) => {
      await queryClient.cancelQueries({ queryKey: notificationKeys.all });

      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all
      );

      // Check if the notification to be deleted is unread
      let isUnread = false;
      if (previousNotifications) {
        const notification = previousNotifications.find(
          (n) => n.id === notificationId
        );
        isUnread = notification ? notification.readAt === null : false;
      }

      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread()
      );

      // Optimistically remove the notification
      if (previousNotifications) {
        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all,
          previousNotifications.filter(
            (notification) => notification.id !== notificationId
          )
        );
      }

      // If the deleted notification was unread, decrement the unread count
      if (
        isUnread &&
        previousUnreadCount !== undefined &&
        previousUnreadCount > 0
      ) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(),
          previousUnreadCount - 1
        );
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

      Alert.alert(
        "Failed to delete notification",
        error.message || "Please try again",
        [
          {
            text: "Retry",
            onPress: () => mutation.mutate(notificationId),
          },
          { text: "Cancel", style: "cancel" },
        ]
      );
    },

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.all });
      queryClient.invalidateQueries({ queryKey: notificationKeys.unread() });
    },
  });

  return mutation;
};
