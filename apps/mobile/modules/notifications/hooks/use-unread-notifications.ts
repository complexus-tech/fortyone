import { useQuery } from "@tanstack/react-query";
import { getUnreadNotifications } from "../queries/get-unread";
import { notificationKeys } from "@/constants/keys";

export const useUnreadNotifications = () => {
  return useQuery({
    queryKey: notificationKeys.unread(),
    queryFn: getUnreadNotifications,
  });
};
