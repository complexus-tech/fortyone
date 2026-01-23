import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { SummaryFilters } from "@/modules/summary/types";
import { getPrioritySummary } from "../queries/analytics/get-priority-summary";

export const usePrioritySummary = (filters?: SummaryFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: ["priority-summary", workspaceSlug, filters],
    queryFn: () =>
      getPrioritySummary({ session: session!, workspaceSlug }, filters),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
