import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { notificationKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getNotifications } from "../queries/get-notifications";

export const useNotifications = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: notificationKeys.all,
    queryFn: () => getNotifications(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
