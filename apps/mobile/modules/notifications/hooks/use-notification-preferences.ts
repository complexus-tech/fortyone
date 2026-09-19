import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { notificationKeys } from "@/constants/keys";
import { updateNotificationPreferences } from "../actions/update-preferences";
import { getNotificationPreferences } from "../queries/get-preferences";
import type { NotificationPreferences, NotificationType } from "../types";

type UpdatePushPreference = { type: NotificationType; enabled: boolean };

export const useNotificationPreferences = () =>
  useQuery({
    queryKey: notificationKeys.preferences(),
    queryFn: getNotificationPreferences,
    staleTime: 10 * 60 * 1_000,
  });

export const useUpdatePushPreference = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ type, enabled }: UpdatePushPreference) =>
      updateNotificationPreferences(type, { pushEnabled: enabled }),
    onMutate: async ({ type, enabled }) => {
      const queryKey = notificationKeys.preferences();
      await queryClient.cancelQueries({ queryKey });
      const previous =
        queryClient.getQueryData<NotificationPreferences>(queryKey);
      if (previous) {
        const current = previous.preferences[type] ?? {
          email: true,
          inApp: true,
          push: true,
        };
        queryClient.setQueryData<NotificationPreferences>(queryKey, {
          ...previous,
          preferences: {
            ...previous.preferences,
            [type]: { ...current, push: enabled },
          },
        });
      }
      return { previous };
    },
    onError: (_error, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData(
          notificationKeys.preferences(),
          context.previous,
        );
      }
    },
    onSettled: () =>
      queryClient.invalidateQueries({
        queryKey: notificationKeys.preferences(),
      }),
  });
};
