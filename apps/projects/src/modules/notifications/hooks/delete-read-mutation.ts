import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { notificationKeys } from "@/constants/keys";
import { deleteReadNotifications } from "../actions/delete-read";
import type { AppNotification } from "../types";

export const useDeleteReadMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: () => deleteReadNotifications(workspaceSlug),

    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: notificationKeys.all(workspaceSlug) });

      const previousNotifications = queryClient.getQueryData<AppNotification[]>(
        notificationKeys.all(workspaceSlug),
      );

      // Optimistically remove read notifications
      if (previousNotifications) {
        const unreadNotifications = previousNotifications.filter(
          (notification) => notification.readAt === null,
        );

        queryClient.setQueryData<AppNotification[]>(
          notificationKeys.all(workspaceSlug),
          unreadNotifications,
        );
      }

      return { previousNotifications };
    },

    onError: (error, _, context) => {
      if (context?.previousNotifications) {
        queryClient.setQueryData(
          notificationKeys.all(workspaceSlug),
          context.previousNotifications,
        );
      }

      toast.error("Failed to delete read notifications", {
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
    },
  });

  return mutation;
};
