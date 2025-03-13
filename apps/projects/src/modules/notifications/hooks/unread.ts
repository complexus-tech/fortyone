import { useQuery } from "@tanstack/react-query";
import { notificationKeys } from "@/constants/keys";
import { getUnreadNotifications } from "../queries/get-unread";

export const useUnreadNotifications = () => {
  return useQuery({
    queryKey: notificationKeys.unread(),
    queryFn: getUnreadNotifications,
  });
};
