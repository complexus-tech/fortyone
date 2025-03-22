import { useQuery } from "@tanstack/react-query";
import { notificationKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getNotificationPreferences } from "../queries/get-preferences";

export const useNotificationPreferences = () => {
  return useQuery({
    queryKey: notificationKeys.preferences(),
    queryFn: () => getNotificationPreferences(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
