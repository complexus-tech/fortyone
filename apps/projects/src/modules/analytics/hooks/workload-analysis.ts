import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getWorkloadAnalysis } from "../queries/get-workload-analysis";
import type { AnalyticsFilters } from "../types";

export const useWorkloadAnalysis = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: analyticsKeys.workloadAnalysis(workspaceSlug, filters),
    queryFn: () =>
      getWorkloadAnalysis({ session: session!, workspaceSlug }, filters),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
