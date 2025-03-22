import { useQuery } from "@tanstack/react-query";
import { notificationKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getNotifications } from "../queries/get-notifications";

export const useNotifications = () => {
  return useQuery({
    queryKey: notificationKeys.all,
    queryFn: () => getNotifications(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
