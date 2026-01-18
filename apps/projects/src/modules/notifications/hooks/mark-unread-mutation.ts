import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { notificationKeys } from "@/constants/keys";
import { markUnread } from "../actions/mark-unread";
import type { AppNotification } from "../types";

export const useMarkUnreadMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (notificationId: string) => markUnread(notificationId, workspaceSlug),

    onMutate: async (notificationId) => {
      await queryClient.cancelQueries({ queryKey: notificationKeys.all(workspaceSlug) });

      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all(workspaceSlug),
      );
      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread(workspaceSlug),
      );

      // Optimistically update the notification as unread
      if (previousNotifications) {
        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all(workspaceSlug),
          previousNotifications.map((notification) =>
            notification.id === notificationId
              ? { ...notification, readAt: null }
              : notification,
          ),
        );

        // Increment the unread count since we're marking a read notification as unread
        if (previousUnreadCount !== undefined) {
          queryClient.setQueryData<number>(
            notificationKeys.unread(workspaceSlug),
            previousUnreadCount + 1,
          );
        }
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

      if (context?.previousUnreadCount) {
        queryClient.setQueryData<number>(
          notificationKeys.unread(workspaceSlug),
          context.previousUnreadCount,
        );
      }

      toast.error("Failed to mark notification as unread", {
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
