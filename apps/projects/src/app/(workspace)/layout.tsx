import type { ReactNode } from "react";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import {
  teamKeys,
  statusKeys,
  workspaceKeys,
  sprintKeys,
} from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { WalkthroughIntegration } from "@/components/walkthrough/walkthrough-integration";
import { getRunningSprints } from "@/modules/sprints/queries/get-running-sprints";
import { Chat } from "@/components/ui/chat";
import { ChatProvider } from "@/context/chat-context";
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
    }),
    queryClient.prefetchQuery({
      queryKey: statusKeys.lists(),
      queryFn: () => getStatuses(session!),
    }),
    queryClient.prefetchQuery({
      queryKey: objectiveKeys.statuses(),
      queryFn: () => getObjectiveStatuses(session!),
    }),
    queryClient.prefetchQuery({
      queryKey: workspaceKeys.lists(),
      queryFn: () => getWorkspaces(token),
    }),
    queryClient.prefetchQuery({
      queryKey: sprintKeys.running(),
      queryFn: () => getRunningSprints(session!),
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ChatProvider>
        {children}
        <Chat />
        <ServerSentEvents />
        <IdentifyUser />
        <WalkthroughIntegration />
      </ChatProvider>
    </HydrationBoundary>
  );
}
