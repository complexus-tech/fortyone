import type { ReactNode } from "react";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import { teamKeys, statusKeys, subscriptionKeys } from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getSubscription } from "@/lib/queries/subscriptions/get-subscription";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { fetchNonCriticalImportantQueries } from "./non-critical-important-queries";

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();

  const headersList = await headers();
  const host = headersList.get("host");
  const subdomain = host?.split(".")[0];
  const token = session?.token || "";
  const workspaces = session?.workspaces || [];

  if (workspaces.length === 0) {
    redirect("https://www.complexus.app/onboarding/create");
  }

  // First try to find workspace in session
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === subdomain?.toLowerCase(),
  );

  if (!workspace) {
    redirect("/unauthorized");
  }
  // kick off non-critical important queries without waiting for them
  const queryClient = fetchNonCriticalImportantQueries(getQueryClient(), token);
  // await critical queries
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: teamKeys.lists(),
      queryFn: () => getTeams(),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
    queryClient.prefetchQuery({
      queryKey: statusKeys.lists(),
      queryFn: () => getStatuses(),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
    queryClient.prefetchQuery({
      queryKey: objectiveKeys.statuses(),
      queryFn: () => getObjectiveStatuses(),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
    queryClient.prefetchQuery({
      queryKey: subscriptionKeys.details,
      queryFn: () => getSubscription(),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      {children}
    </HydrationBoundary>
  );
}
