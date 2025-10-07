import { useQuery } from "@tanstack/react-query";
import { subscriptionKeys } from "@/constants/keys";
import { getSubscription } from "@/lib/queries/get-subscription";

export const useSubscription = () => {
  return useQuery({
    queryKey: subscriptionKeys.details,
    queryFn: getSubscription,
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchOnMount: true,
    refetchOnReconnect: true,
  });
};
