import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getTeamPerformance } from "../queries/get-team-performance";
import type { AnalyticsFilters } from "../types";

export const useTeamPerformance = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: analyticsKeys.teamPerformance(filters),
    queryFn: () => getTeamPerformance(filters, session!),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
