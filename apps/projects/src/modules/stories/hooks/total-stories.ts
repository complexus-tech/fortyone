import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { storyKeys } from "../constants";
import { getTotalStories } from "../queries/get-total-stories";

export const useTotalStories = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { tier } = useSubscriptionFeatures();

  return useQuery({
    queryKey: storyKeys.total(workspaceSlug),
    queryFn: () => getTotalStories({ session: session!, workspaceSlug }),
    enabled: tier === "free",
    staleTime: Number(DURATION_FROM_MILLISECONDS.MINUTE),
  });
};
