import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getTeamPerformance } from "../queries/get-team-performance";
import type { AnalyticsFilters } from "../types";

export const useTeamPerformance = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: analyticsKeys.teamPerformance(workspaceSlug, filters),
    queryFn: () => getTeamPerformance({ session: session!, workspaceSlug }, filters),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
