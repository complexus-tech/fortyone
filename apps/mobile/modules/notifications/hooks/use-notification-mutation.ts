import { useQueryClient } from "@tanstack/react-query";
import { useSessionMutation } from "@/lib/use-session-mutation";
import { toast } from "sonner-native";
import { notificationKeys } from "@/constants/keys";
import { updateNotificationsCache } from "../utils/cache";
import type { NotificationChange, NotificationsData } from "../utils/cache";

export const useNotificationMutation = <TVariables>({
  mutationFn,
  change,
  errorTitle,
  successTitle,
  optimistic = true,
}: {
  mutationFn: (variables: TVariables) => Promise<unknown>;
  change: (variables: TVariables) => NotificationChange;
  errorTitle: string;
  successTitle?: string;
  optimistic?: boolean;
}) => {
  const client = useQueryClient();
  const allKey = notificationKeys.all;
  const listKey = notificationKeys.lists();
  const unreadKey = notificationKeys.unread();

  return useSessionMutation({
    mutationFn,
    onMutate: async (variables) => {
      await client.cancelQueries({ queryKey: allKey });
      const before = client.getQueryData<NotificationsData>(listKey);
      const beforeUnread = client.getQueryData<number>(unreadKey);
      if (!optimistic) return undefined;
      const next = updateNotificationsCache(
        before,
        beforeUnread,
        change(variables),
      );
      if (next.data) client.setQueryData(listKey, next.data);
      if (next.unreadCount !== undefined)
        client.setQueryData(unreadKey, next.unreadCount);
      return {
        before,
        beforeUnread,
        after: client.getQueryData(listKey),
        afterUnread: next.unreadCount,
      };
    },
    onError: (error, _variables, snapshot) => {
      // Do not overwrite another mutation that completed after this snapshot.
      if (snapshot && client.getQueryData(listKey) === snapshot.after) {
        client.setQueryData(listKey, snapshot.before);
      }
      if (snapshot && client.getQueryData(unreadKey) === snapshot.afterUnread) {
        client.setQueryData(unreadKey, snapshot.beforeUnread);
      }
      toast.error(errorTitle, { description: error.message });
    },
    onSuccess: () => {
      if (successTitle) toast.success(successTitle);
    },
    onSettled: () => client.invalidateQueries({ queryKey: allKey }),
  });
};
