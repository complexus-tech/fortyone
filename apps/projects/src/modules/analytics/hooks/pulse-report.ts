import { useQuery } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useSession } from "@/lib/auth/client";
import { getPulseReport } from "../queries/get-pulse-report";
import type { AnalyticsFilters } from "../types";

export const usePulseReport = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: analyticsKeys.pulseReport(workspaceSlug, filters),
    queryFn: () =>
      getPulseReport({ session: session!, workspaceSlug }, filters),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
