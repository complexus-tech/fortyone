import { useQuery } from "@tanstack/react-query";
import { notificationKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getUnreadNotifications } from "../queries/get-unread";

export const useUnreadNotifications = () => {
  return useQuery({
    queryKey: notificationKeys.unread(),
    queryFn: () => getUnreadNotifications(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
