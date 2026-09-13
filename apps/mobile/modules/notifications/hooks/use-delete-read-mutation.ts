import { deleteReadNotifications } from "../actions/delete-read";
import { useNotificationMutation } from "./use-notification-mutation";

export const useDeleteReadMutation = () =>
  useNotificationMutation<void>({
    mutationFn: deleteReadNotifications,
    change: () => ({ type: "delete-read" }),
    errorTitle: "Failed to delete read notifications",
    successTitle: "Read notifications deleted",
  });
