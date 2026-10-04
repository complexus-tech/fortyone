import { readNotification } from "../actions/read";
import { useNotificationMutation } from "./use-notification-mutation";
import type { ReadNotificationInput } from "lib/src/notification-read";

export const useReadNotificationMutation = (isOptimistic = true) =>
  useNotificationMutation<ReadNotificationInput>({
    mutationFn: ({ id, observedCreatedAt }) =>
      readNotification(id, observedCreatedAt),
    change: ({ id, observedCreatedAt }) => ({
      type: "read",
      id,
      observedCreatedAt,
    }),
    errorTitle: "Failed to mark notification as read",
    optimistic: isOptimistic,
  });
