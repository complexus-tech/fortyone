import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { notificationKeys } from "@/constants/keys";
import { deleteNotification } from "../actions/delete";
import type { AppNotification } from "../types";

export const useDeleteMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (notificationId: string) => deleteNotification(notificationId, workspaceSlug),

    onMutate: async (notificationId) => {
      await queryClient.cancelQueries({ queryKey: notificationKeys.all(workspaceSlug) });

      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all(workspaceSlug),
      );

      // Check if the notification to be deleted is unread
      let isUnread = false;
      if (previousNotifications) {
        const notification = previousNotifications.find(
          (n) => n.id === notificationId,
        );
        isUnread = notification ? notification.readAt === null : false;
      }

      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread(workspaceSlug),
      );

      // Optimistically remove the notification
      if (previousNotifications) {
        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all(workspaceSlug),
          previousNotifications.filter(
            (notification) => notification.id !== notificationId,
          ),
        );
      }

      // If the deleted notification was unread, decrement the unread count
      if (
        isUnread &&
        previousUnreadCount !== undefined &&
        previousUnreadCount > 0
      ) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(workspaceSlug),
          previousUnreadCount - 1,
        );
      }

      return { previousNotifications, previousUnreadCount };
    },

    onError: (error, notificationId, context) => {
      if (context?.previousNotifications) {
        queryClient.setQueryData(
          notificationKeys.all(workspaceSlug),
          context.previousNotifications,
        );
      }

      if (context?.previousUnreadCount !== undefined) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(workspaceSlug),
          context.previousUnreadCount,
        );
      }

      toast.error("Failed to delete notification", {
        description: error.message || "Please try again",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(notificationId);
          },
        },
      });
    },

    onSuccess: (res) => {
      if (res?.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: notificationKeys.all(workspaceSlug) });
      queryClient.invalidateQueries({ queryKey: notificationKeys.unread(workspaceSlug) });
    },
  });

  return mutation;
};
