import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Alert } from "react-native";
import { notificationKeys } from "@/constants/keys";
import { deleteReadNotifications } from "../actions/delete-read";
import type { AppNotification } from "../types";

export const useDeleteReadMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: deleteReadNotifications,

    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: notificationKeys.all });

      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all
      );

      // Optimistically remove read notifications
      if (previousNotifications) {
        const unreadNotifications = previousNotifications.filter(
          (notification) => notification.readAt === null
        );

        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all,
          unreadNotifications
        );
      }

      return { previousNotifications };
    },

    onError: (error, _, context) => {
      if (context?.previousNotifications) {
        queryClient.setQueryData(
          notificationKeys.all,
          context.previousNotifications
        );
      }

      Alert.alert(
        "Failed to delete read notifications",
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
