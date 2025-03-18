import { useSuspenseQuery } from "@tanstack/react-query";
import { notificationKeys } from "@/constants/keys";
import { getNotificationPreferences } from "../queries/get-preferences";

export const useNotificationPreferences = () => {
  return useSuspenseQuery({
    queryKey: notificationKeys.preferences(),
    queryFn: getNotificationPreferences,
  });
};
