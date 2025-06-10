import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getTimelineTrends } from "../queries/get-timeline-trends";
import type { AnalyticsFilters } from "../types";

export const useTimelineTrends = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: analyticsKeys.timelineTrends(filters),
    queryFn: () => getTimelineTrends(filters, session!),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
