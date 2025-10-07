import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Alert } from "react-native";
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

  const mutation = useMutation({
    mutationFn: (notificationId: string) => readNotification(notificationId),

    onMutate: async (notificationId) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: notificationKeys.all });

      // Get the previous data
      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all
      );

      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread()
      );

      if (previousUnreadCount && isOptimistic) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(),
          previousUnreadCount - 1
        );
      }

      // Optimistically update the notifications
      if (previousNotifications && isOptimistic) {
        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all,
          previousNotifications.map((notification) =>
            notification.id === notificationId
              ? { ...notification, readAt: new Date().toISOString() }
              : notification
          )
        );
      }

      return { previousNotifications, previousUnreadCount };
    },

    onError: (error, notificationId, context) => {
      // Rollback on error
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
        "Failed to mark notification as read",
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
