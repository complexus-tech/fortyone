import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getWorkspaceOverview } from "../queries/get-workspace-overview";
import type { AnalyticsFilters } from "../types";

export const useWorkspaceOverview = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: analyticsKeys.overview(filters),
    queryFn: () => getWorkspaceOverview(filters, session!),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
