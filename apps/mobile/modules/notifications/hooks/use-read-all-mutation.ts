import { readAllNotifications } from "../actions/read-all";
import { useNotificationMutation } from "./use-notification-mutation";

export const useReadAllNotificationsMutation = () =>
  useNotificationMutation<void>({
    mutationFn: readAllNotifications,
    change: () => ({ type: "read-all" }),
    errorTitle: "Failed to mark all notifications as read",
    successTitle: "All notifications marked as read",
  });
