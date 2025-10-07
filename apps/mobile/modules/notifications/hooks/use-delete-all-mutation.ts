import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Alert } from "react-native";
import { notificationKeys } from "@/constants/keys";
import { deleteAllNotifications } from "../actions/delete-all";
import type { AppNotification } from "../types";

export const useDeleteAllMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: deleteAllNotifications,

    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: notificationKeys.all });

      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all
      );
      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread()
      );

      // Optimistically clear all notifications
      queryClient.setQueryData<AppNotification[]>(notificationKeys.all, []);
      queryClient.setQueryData(notificationKeys.unread(), 0);

      return { previousNotifications, previousUnreadCount };
    },

    onError: (error, _, context) => {
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
        "Failed to delete all notifications",
        error.message || "Please try again",
        [
          {
            text: "Retry",
            onPress: () => mutation.mutate(),
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
