import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStoryAnalytics } from "../queries/get-story-analytics";
import type { AnalyticsFilters } from "../types";

export const useStoryAnalytics = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: analyticsKeys.storyAnalytics(filters),
    queryFn: () => getStoryAnalytics(filters, session!),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
