import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { notificationKeys } from "@/constants/keys";
import { deleteAllNotifications } from "../actions/delete-all";
import type { AppNotification } from "../types";

export const useDeleteAllMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: () => deleteAllNotifications(workspaceSlug),

    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: notificationKeys.all(workspaceSlug) });

      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all(workspaceSlug),
      );
      const previousUnreadCount = queryClient.getQueryData<number>(
        notificationKeys.unread(workspaceSlug),
      );

      // Optimistically clear all notifications
      queryClient.setQueryData<AppNotification[]>(notificationKeys.all(workspaceSlug), []);
      queryClient.setQueryData(notificationKeys.unread(workspaceSlug), 0);

      return { previousNotifications, previousUnreadCount };
    },

    onError: (error, _, context) => {
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

      toast.error("Failed to delete all notifications", {
        description: error.message || "Please try again",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate();
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
