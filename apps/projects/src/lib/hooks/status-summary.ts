import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { SummaryFilters } from "@/modules/summary/types";
import { getStatusSummary } from "../queries/analytics/get-status-summary";

export const useStatusSummary = (filters?: SummaryFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: ["status-summary", workspaceSlug, filters],
    queryFn: () =>
      getStatusSummary({ session: session!, workspaceSlug }, filters),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
