import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getPrioritySummary } from "../queries/analytics/get-priority-summary";

export const usePrioritySummary = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: ["priority-summary"],
    queryFn: () => getPrioritySummary(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
