import { readNotification } from "../actions/read";
import { useNotificationMutation } from "./use-notification-mutation";

export const useReadNotificationMutation = (isOptimistic = true) =>
  useNotificationMutation<string>({
    mutationFn: readNotification,
    change: (id) => ({ type: "read", id }),
    errorTitle: "Failed to mark notification as read",
    optimistic: isOptimistic,
  });
