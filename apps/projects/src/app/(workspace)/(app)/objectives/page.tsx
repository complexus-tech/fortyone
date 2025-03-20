import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { ObjectivesList } from "@/modules/objectives";
import { getQueryClient } from "@/app/get-query-client";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { objectiveKeys } from "@/modules/objectives/constants";

export const metadata: Metadata = {
  title: "Objectives",
};

export default function Page() {
  const queryClient = getQueryClient();
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.all,
    queryFn: () => getObjectives(),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ObjectivesList />
    </HydrationBoundary>
  );
}
