import type { ReactNode } from "react";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { redirect } from "next/navigation";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import {
  teamKeys,
  statusKeys,
  workspaceKeys,
  sprintKeys,
  userKeys,
} from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { WalkthroughIntegration } from "@/components/walkthrough/walkthrough-integration";
import { getRunningSprints } from "@/modules/sprints/queries/get-running-sprints";
import { Chat } from "@/components/ui/chat";
import { ChatProvider } from "@/context/chat-context";
import { switchWorkspace } from "@/lib/actions/users/switch-workspace";
import { getProfile } from "@/lib/queries/users/profile";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { ServerSentEvents } from "../server-sent-events";
import { fetchNonCriticalImportantQueries } from "./non-critical-important-queries";
import { IdentifyUser } from "./identify";

export default async function RootLayout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ workspaceSlug: string }>;
}) {
  const session = await auth();

  if (!session) {
    redirect("/");
  }
  const { workspaceSlug } = await params;
  const ctx = { session: session!, workspaceSlug };
  const workspaces = await getWorkspaces(session?.token);

  if (workspaces.length === 0) {
    redirect("/onboarding/create");
  }

  // First try to find workspace
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === workspaceSlug?.toLowerCase(),
  );

  if (!workspace) {
    redirect("/unauthorized");
  }

  // kick off non-critical important queries without waiting for them
  const queryClient = fetchNonCriticalImportantQueries(getQueryClient(), ctx);

  // await critical queries
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: teamKeys.lists(workspaceSlug),
      queryFn: () => getTeams(ctx),
    }),
    queryClient.prefetchQuery({
      queryKey: statusKeys.lists(workspaceSlug),
      queryFn: () => getStatuses(ctx),
    }),
    queryClient.prefetchQuery({
      queryKey: objectiveKeys.statuses(workspaceSlug),
      queryFn: () => getObjectiveStatuses(ctx),
    }),
    queryClient.prefetchQuery({
      queryKey: workspaceKeys.lists(),
      queryFn: () => getWorkspaces(ctx.session.token),
    }),
    queryClient.prefetchQuery({
      queryKey: sprintKeys.running(workspaceSlug),
      queryFn: () => getRunningSprints(ctx),
    }),
    queryClient.prefetchQuery({
      queryKey: userKeys.profile(),
      queryFn: () => getProfile(ctx.session),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
    // switchWorkspace(workspace.id),
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
