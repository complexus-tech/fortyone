import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { notificationKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getNotificationPreferences } from "../queries/get-preferences";

export const useNotificationPreferences = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: notificationKeys.preferences(),
    queryFn: () => getNotificationPreferences(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
