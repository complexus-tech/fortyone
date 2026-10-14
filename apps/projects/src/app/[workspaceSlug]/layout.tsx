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
import { getStatuses } from "@/lib/queries/states/get-states";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getRunningSprints } from "@/modules/sprints/queries/get-running-sprints";
import { Chat } from "@/components/ui/chat";
import { WalkthroughProvider } from "@/components/walkthrough/walkthrough-provider";
import { WORKSPACE_GETTING_STARTED_TOUR_VERSION } from "@/components/walkthrough/walkthrough-tour";
import { ChatProvider } from "@/context/chat-context";
import { getProfile } from "@/lib/queries/users/profile";
import { getCookieHeader } from "@/lib/http/header";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getWorkspaceAccess } from "@/modules/auth/workspace-access";
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
  const [session, cookieHeader, { workspaceSlug }] = await Promise.all([
    auth(),
    getCookieHeader(),
    params,
  ]);
  const ctx = { session, workspaceSlug, cookieHeader };
  const workspaces = await getWorkspaces(session?.token, cookieHeader);

  if (workspaces.length === 0) {
    redirect(session ? "/onboarding/create" : "/");
  }

  // First try to find workspace
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === workspaceSlug.toLowerCase(),
  );

  if (!workspace) {
    redirect("/unauthorized");
  }

  const access = await getWorkspaceAccess(ctx);
  if (access === "sso-required") {
    redirect(`/auth/sso/${encodeURIComponent(workspace.slug)}`);
  }
  if (access === "denied") redirect("/unauthorized");

  const queryClient = getQueryClient();
  // Reuse the membership list already checked above within this request only.
  // Hydrating it avoids a second lookup before the workspace switcher is ready.
  queryClient.setQueryData(workspaceKeys.lists(), workspaces);
  fetchNonCriticalImportantQueries(
    queryClient,
    ctx,
    session?.user.id ?? "anonymous",
  );

  // Settle the data used by the initial navigation and terminology. Supporting
  // feature queries stream independently through the hydration boundary.
  await Promise.all([
    // Terminology affects the shell's server-rendered text and must hydrate
    // from the same settled snapshot on the client.
    queryClient.prefetchQuery({
      queryKey: workspaceKeys.settings(workspaceSlug),
      queryFn: () => getWorkspaceSettings(ctx),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
    queryClient.prefetchQuery({
      queryKey: teamKeys.lists(workspaceSlug),
      queryFn: () => getTeams(ctx),
    }),
    queryClient.prefetchQuery({
      queryKey: statusKeys.lists(workspaceSlug),
      queryFn: () => getStatuses(ctx),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
    }),
    queryClient.prefetchQuery({
      queryKey: sprintKeys.running(workspaceSlug),
      queryFn: () => getRunningSprints(ctx),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
    }),
    queryClient.prefetchQuery({
      queryKey: userKeys.profile(),
      queryFn: () =>
        getProfile({
          token: ctx.session?.token,
          cookieHeader: ctx.cookieHeader,
        }),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <WalkthroughProvider
        autoStart
        version={WORKSPACE_GETTING_STARTED_TOUR_VERSION}
      >
        <ChatProvider key={workspaceSlug}>
          {children}
          <ServerSentEvents />
          <IdentifyUser />
          <Chat />
        </ChatProvider>
      </WalkthroughProvider>
    </HydrationBoundary>
  );
}
