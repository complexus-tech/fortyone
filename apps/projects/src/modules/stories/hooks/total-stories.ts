import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { storyKeys } from "../constants";
import { getTotalStories } from "../queries/get-total-stories";

export const useTotalStories = () => {
  const { data: session } = useSession();
  const { tier } = useSubscriptionFeatures();
  return useQuery({
    queryKey: storyKeys.total(),
    queryFn: () => getTotalStories(session!),
    enabled: tier === "free",
    staleTime: Number(DURATION_FROM_MILLISECONDS.MINUTE),
  });
};
