import { useSuspenseQuery } from "@tanstack/react-query";
import { getSummary } from "../queries/analytics/get-summary";

export const useSummary = () => {
  return useSuspenseQuery({
    queryKey: ["summary"],
    queryFn: () => getSummary(),
  });
};
