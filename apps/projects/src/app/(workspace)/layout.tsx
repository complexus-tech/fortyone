import type { ReactNode } from "react";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import { teamKeys, statusKeys, workspaceKeys } from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { getStatuses } from "@/lib/queries/states/get-states";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { ServerSentEvents } from "../server-sent-events";
import { fetchNonCriticalImportantQueries } from "./non-critical-important-queries";
import { IdentifyUser } from "./identify";

const isLocalhost = process.env.NODE_ENV === "development";

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
    redirect(
      isLocalhost
        ? "https://complexus.lc/onboarding/create"
        : "https://www.complexus.app/onboarding/create",
    );
  }

  // First try to find workspace in session
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === subdomain?.toLowerCase(),
  );

  if (!workspace) {
    redirect("/unauthorized");
  }
  // kick off non-critical important queries without waiting for them
  const queryClient = fetchNonCriticalImportantQueries(
    getQueryClient(),
    session!,
  );
  // await critical queries
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: teamKeys.lists(),
      queryFn: () => getTeams(session!),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
    queryClient.prefetchQuery({
      queryKey: statusKeys.lists(),
      queryFn: () => getStatuses(session!),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
    queryClient.prefetchQuery({
      queryKey: objectiveKeys.statuses(),
      queryFn: () => getObjectiveStatuses(session!),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
    queryClient.prefetchQuery({
      queryKey: workspaceKeys.lists(),
      queryFn: () => getWorkspaces(token),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      {children}
      <ServerSentEvents />
      <IdentifyUser />
    </HydrationBoundary>
  );
}
