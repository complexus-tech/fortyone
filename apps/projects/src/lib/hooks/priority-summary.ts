import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getPrioritySummary } from "../queries/analytics/get-priority-summary";

export const usePrioritySummary = () => {
  return useQuery({
    queryKey: ["priority-summary"],
    queryFn: () => getPrioritySummary(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
