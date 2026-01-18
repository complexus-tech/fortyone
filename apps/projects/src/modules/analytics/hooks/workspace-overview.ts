import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getWorkspaceOverview } from "../queries/get-workspace-overview";
import type { AnalyticsFilters } from "../types";

export const useWorkspaceOverview = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: analyticsKeys.overview(workspaceSlug, filters),
    queryFn: () => getWorkspaceOverview({ session: session!, workspaceSlug }, filters),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
