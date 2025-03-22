import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getSummary } from "../queries/analytics/get-summary";

export const useSummary = () => {
  return useQuery({
    queryKey: ["summary"],
    queryFn: () => getSummary(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
