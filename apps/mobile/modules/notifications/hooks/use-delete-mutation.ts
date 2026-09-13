import { deleteNotification } from "../actions/delete";
import { useNotificationMutation } from "./use-notification-mutation";

export const useDeleteMutation = () =>
  useNotificationMutation<string>({
    mutationFn: deleteNotification,
    change: (id) => ({ type: "delete", id }),
    errorTitle: "Failed to delete notification",
    successTitle: "Notification deleted",
  });
