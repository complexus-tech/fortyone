import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { getSubscription } from "@/lib/queries/subscriptions/get-subscription";
import { subscriptionKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useSubscription = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: subscriptionKeys.details,
    queryFn: () => getSubscription(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    refetchOnMount: true,
  });
};
