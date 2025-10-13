import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner-native";
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

      toast.error("Failed to delete read notifications", {
        description: error.message || "Please try again",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(),
        },
      });
    },

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.all });
      queryClient.invalidateQueries({ queryKey: notificationKeys.unread() });
      toast.success("Read notifications deleted");
    },
  });

  return mutation;
};
