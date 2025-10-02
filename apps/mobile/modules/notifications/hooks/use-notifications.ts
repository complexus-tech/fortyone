import { useQuery } from "@tanstack/react-query";
import { getNotifications } from "../queries/get-notifications";
import { notificationKeys } from "@/constants/keys";

export const useNotifications = () => {
  return useQuery({
    queryKey: notificationKeys.all,
    queryFn: getNotifications,
  });
};
