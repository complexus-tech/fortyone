import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { SummaryFilters } from "@/modules/summary/types";
import { getContributions } from "../queries/analytics/get-contributions";

export const useContributions = (filters?: SummaryFilters) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: ["contributions", workspaceSlug, filters],
    queryFn: () =>
      getContributions({ session: session!, workspaceSlug }, filters),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
