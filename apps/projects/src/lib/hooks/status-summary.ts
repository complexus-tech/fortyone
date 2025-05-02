import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStatusSummary } from "../queries/analytics/get-status-summary";

export const useStatusSummary = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: ["status-summary"],
    queryFn: () => getStatusSummary(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
