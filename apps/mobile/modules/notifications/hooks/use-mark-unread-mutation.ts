import { markUnread } from "../actions/mark-unread";
import { useNotificationMutation } from "./use-notification-mutation";

export const useMarkUnreadMutation = () =>
  useNotificationMutation<string>({
    mutationFn: markUnread,
    change: (id) => ({ type: "unread", id }),
    errorTitle: "Failed to mark notification as unread",
    successTitle: "Notification marked as unread",
  });
