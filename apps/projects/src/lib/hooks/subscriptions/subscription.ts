import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { getSubscription } from "@/lib/queries/subscriptions/get-subscription";
import { subscriptionKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useSubscription = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: subscriptionKeys.details(workspaceSlug),
    queryFn: () => getSubscription({ session: session!, workspaceSlug }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    refetchInterval: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    refetchIntervalInBackground: true,
    refetchOnMount: true,
    refetchOnReconnect: true,
  });
};
