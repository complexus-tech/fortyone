import type { ReactNode } from "react";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import { teamKeys, statusKeys } from "@/constants/keys";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { getStatuses } from "@/lib/queries/states/get-states";
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

  // First try to find workspace in session
  let workspace = session?.workspaces.find(
    (w) => w.slug.toLowerCase() === subdomain?.toLowerCase(),
  );

  // Only fetch workspaces if not found in session
  let workspaces = session?.workspaces || [];
  if (!workspace) {
    workspaces = await getWorkspaces(token);
    workspace = workspaces.find(
      (w) => w.slug.toLowerCase() === subdomain?.toLowerCase(),
    );
  }

  if (workspaces.length === 0) {
    redirect("/onboarding/create");
  }

  if (!workspace) {
    redirect("/unauthorized");
  }
  // kick off non-critical important queries without waiting for them
  const queryClient = fetchNonCriticalImportantQueries(getQueryClient(), token);
  // await critical queries
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: teamKeys.lists(),
      queryFn: getTeams,
    }),
    queryClient.prefetchQuery({
      queryKey: statusKeys.lists(),
      queryFn: getStatuses,
    }),
    queryClient.prefetchQuery({
      queryKey: objectiveKeys.statuses(),
      queryFn: getObjectiveStatuses,
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      {children}
    </HydrationBoundary>
  );
}
