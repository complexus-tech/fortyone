import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { notificationKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getUnreadNotifications } from "../queries/get-unread";

export const useUnreadNotifications = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: notificationKeys.unread(),
    queryFn: () => getUnreadNotifications(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
