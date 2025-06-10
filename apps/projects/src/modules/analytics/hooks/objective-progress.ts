import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { analyticsKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getObjectiveProgress } from "../queries/get-objective-progress";
import type { AnalyticsFilters } from "../types";

export const useObjectiveProgress = (filters?: AnalyticsFilters) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: analyticsKeys.objectiveProgress(filters),
    queryFn: () => getObjectiveProgress(filters, session!),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
