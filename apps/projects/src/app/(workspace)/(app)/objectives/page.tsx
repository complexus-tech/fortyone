import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { ObjectivesList } from "@/modules/objectives";
import { getQueryClient } from "@/app/get-query-client";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { objectiveKeys } from "@/modules/objectives/constants";
import { auth } from "@/auth";

export const metadata: Metadata = {
  title: "Objectives",
};

export default async function Page() {
  const queryClient = getQueryClient();
  const session = await auth();
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.all,
    queryFn: () => getObjectives(session!),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ObjectivesList />
    </HydrationBoundary>
  );
}
