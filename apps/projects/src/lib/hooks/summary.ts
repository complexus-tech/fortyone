import { useQuery } from "@tanstack/react-query";
import { getSummary } from "../queries/analytics/get-summary";

export const useSummary = () => {
  return useQuery({
    queryKey: ["summary"],
    queryFn: getSummary,
  });
};
