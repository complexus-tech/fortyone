import { useQuery } from "@tanstack/react-query";
import { getSubscription } from "@/lib/queries/subscriptions/get-subscription";
import { subscriptionKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useSubscription = () => {
  return useQuery({
    queryKey: subscriptionKeys.details,
    queryFn: () => getSubscription(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
