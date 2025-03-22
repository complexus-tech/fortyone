import { useQuery } from "@tanstack/react-query";
import { notificationKeys } from "@/constants/keys";
import { getNotificationPreferences } from "../queries/get-preferences";

export const useNotificationPreferences = () => {
  return useQuery({
    queryKey: notificationKeys.preferences(),
    queryFn: getNotificationPreferences,
  });
};
