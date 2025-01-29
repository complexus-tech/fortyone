import type { ReactNode } from "react";
import { SessionProvider } from "next-auth/react";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { ApplicationLayout } from "@/components/layouts";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import { getMembers } from "@/lib/queries/members/get-members";
import {
  memberKeys,
  labelKeys,
  teamKeys,
  sprintKeys,
  statusKeys,
  workspaceKeys,
} from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const queryClient = getQueryClient();
  const session = await auth();

  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: statusKeys.lists(),
      queryFn: getStatuses,
    }),
    queryClient.prefetchQuery({
      queryKey: teamKeys.lists(),
      queryFn: getTeams,
    }),
    queryClient.prefetchQuery({
      queryKey: sprintKeys.lists(),
      queryFn: getSprints,
    }),
    queryClient.prefetchQuery({
      queryKey: memberKeys.lists(),
      queryFn: getMembers,
    }),
    queryClient.prefetchQuery({
      queryKey: objectiveKeys.list(),
      queryFn: getObjectives,
    }),
    queryClient.prefetchQuery({
      queryKey: labelKeys.lists(),
      queryFn: () => getLabels(),
    }),
    queryClient.prefetchQuery({
      queryKey: workspaceKeys.detail(),
      queryFn: () => getWorkspace(),
    }),
    queryClient.prefetchQuery({
      queryKey: workspaceKeys.lists(),
      queryFn: () => getWorkspaces(),
    }),
  ]);

  return (
    <SessionProvider session={session}>
      <HydrationBoundary state={dehydrate(queryClient)}>
        <ApplicationLayout>{children}</ApplicationLayout>
      </HydrationBoundary>
    </SessionProvider>
  );
}
