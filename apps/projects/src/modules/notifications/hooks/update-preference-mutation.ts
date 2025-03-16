import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { notificationKeys } from "@/constants/keys";
import { updateNotificationPreferences } from "../actions/update-preferences";
import type {
  NotificationPreferences,
  UpdateNotificationPreferences,
} from "../types";

type NotificationType =
  | "story_update"
  | "objective_update"
  | "comment_reply"
  | "mention"
  | "key_result_update"
  | "story_comment";

type UpdatePreferenceParams = {
  type: NotificationType;
  preferences: UpdateNotificationPreferences;
};

export const useUpdateNotificationPreferenceMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ type, preferences }: UpdatePreferenceParams) =>
      updateNotificationPreferences(preferences, type),

    onMutate: async ({ type, preferences }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({
        queryKey: notificationKeys.preferences(),
      });

      // Get the previous data
      const previousPreferences =
        queryClient.getQueryData<NotificationPreferences>(
          notificationKeys.preferences(),
        );

      // Optimistically update to the new value
      if (previousPreferences) {
        queryClient.setQueryData<NotificationPreferences>(
          notificationKeys.preferences(),
          {
            ...previousPreferences,
            preferences: {
              ...previousPreferences.preferences,
              [type]: {
                email:
                  preferences.emailEnabled ??
                  previousPreferences.preferences[type].email,
                inApp:
                  preferences.inAppEnabled ??
                  previousPreferences.preferences[type].inApp,
              },
            },
          },
        );
      }

      return { previousPreferences };
    },

    onError: (error, variables, context) => {
      // Revert to the previous value
      if (context?.previousPreferences) {
        queryClient.setQueryData(
          notificationKeys.preferences(),
          context.previousPreferences,
        );
      }

      toast.error("Failed to update notification preferences", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (res) => {
      if (res?.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({
        queryKey: notificationKeys.preferences(),
      });
    },
  });

  return mutation;
};
