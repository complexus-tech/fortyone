import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStatusSummary } from "../queries/analytics/get-status-summary";

export const useStatusSummary = () => {
  return useQuery({
    queryKey: ["status-summary"],
    queryFn: () => getStatusSummary(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
