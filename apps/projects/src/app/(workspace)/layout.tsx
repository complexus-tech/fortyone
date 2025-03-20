import type { ReactNode } from "react";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import { teamKeys, statusKeys } from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { getStatuses } from "@/lib/queries/states/get-states";
import { fetchNonCriticalImportantQueries } from "./non-critical-important-queries";

export const dynamic = "force-dynamic";

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

  // redirect to login if not authenticated on local dev only
  // eslint-disable-next-line turbo/no-undeclared-env-vars -- ok for this
  if (!session && process.env.NODE_ENV === "development") {
    redirect("/login");
  }

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
