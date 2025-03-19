import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import type { Metadata } from "next";
import { getQueryClient } from "@/app/get-query-client";
import { SummaryPage } from "@/modules/summary";
import { getActivities } from "@/lib/queries/activities/get-activities";
import { getSummary } from "@/lib/queries/analytics/get-summary";

export const metadata: Metadata = {
  title: "Summary",
};

export default function Page() {
  const queryClient = getQueryClient();
  Promise.all([
    queryClient.prefetchQuery({
      queryKey: ["activities"],
      queryFn: getActivities,
    }),
    queryClient.prefetchQuery({
      queryKey: ["summary"],
      queryFn: getSummary,
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <SummaryPage />
    </HydrationBoundary>
  );
}
