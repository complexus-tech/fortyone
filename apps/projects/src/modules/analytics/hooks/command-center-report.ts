import { useQuery } from "@tanstack/react-query";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { getCommandCenterReport } from "../queries/get-command-center-report";
import type { AnalyticsFilters } from "../types";

export const useCommandCenterReport = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: analyticsKeys.commandCenter(workspaceSlug, filters),
    queryFn: () =>
      getCommandCenterReport({ session: session!, workspaceSlug }, filters),
    enabled: Boolean(session && workspaceSlug),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
