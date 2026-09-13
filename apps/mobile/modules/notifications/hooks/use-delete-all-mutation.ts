import { deleteAllNotifications } from "../actions/delete-all";
import { useNotificationMutation } from "./use-notification-mutation";

export const useDeleteAllMutation = () =>
  useNotificationMutation<void>({
    mutationFn: deleteAllNotifications,
    change: () => ({ type: "delete-all" }),
    errorTitle: "Failed to delete all notifications",
    successTitle: "All notifications deleted",
  });
