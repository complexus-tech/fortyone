import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStoryAnalytics } from "../queries/get-story-analytics";
import type { AnalyticsFilters } from "../types";

export const useStoryAnalytics = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: analyticsKeys.storyAnalytics(workspaceSlug, filters),
    queryFn: () => getStoryAnalytics({ session: session!, workspaceSlug }, filters),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
