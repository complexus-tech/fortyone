import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getContributions } from "../queries/analytics/get-contributions";

export const useContributions = () => {
  return useQuery({
    queryKey: ["contributions"],
    queryFn: () => getContributions(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
