import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { ObjectivesList } from "@/modules/objectives";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getQueryClient } from "@/app/get-query-client";

export const metadata: Metadata = {
  title: "Objectives",
};

export default async function Page() {
  const queryClient = getQueryClient();
  await queryClient.prefetchQuery({
    queryKey: objectiveKeys.list(),
    queryFn: getObjectives,
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ObjectivesList />
    </HydrationBoundary>
  );
}
