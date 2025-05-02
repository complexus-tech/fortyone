import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getSummary } from "../queries/analytics/get-summary";

export const useSummary = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: ["summary"],
    queryFn: () => getSummary(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
