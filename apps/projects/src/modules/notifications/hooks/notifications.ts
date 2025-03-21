import { useSuspenseQuery } from "@tanstack/react-query";
import { notificationKeys } from "@/constants/keys";
import { getNotifications } from "../queries/get-notifications";

export const useNotifications = () => {
  return useSuspenseQuery({
    queryKey: notificationKeys.all,
    queryFn: () => getNotifications(),
  });
};
