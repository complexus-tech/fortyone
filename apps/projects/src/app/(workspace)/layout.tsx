import type { ReactNode } from "react";
import { SessionProvider } from "next-auth/react";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import ky from "ky";
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
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import type { ApiResponse, Workspace } from "@/types";
import { OnlineStatusMonitor } from "../online-monitor";

const apiURL = process.env.NEXT_PUBLIC_API_URL;
const getWorkspaces = async (token?: string) => {
  const workspaces = await ky
    .get(`${apiURL}/workspaces`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .json<ApiResponse<Workspace[]>>();
  return workspaces.data!;
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const queryClient = getQueryClient();
  const session = await auth();
  const headersList = await headers();
  const host = headersList.get("host");
  const subdomain = host?.split(".")[0];
  const workspaces = await getWorkspaces(session?.token);

  if (workspaces.length === 0) {
    redirect("/onboarding/create");
  }

  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === subdomain?.toLowerCase(),
  );

  if (!workspace) {
    redirect("/unauthorized");
  }

  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: statusKeys.lists(),
      queryFn: getStatuses,
    }),
    queryClient.prefetchQuery({
      queryKey: objectiveKeys.statuses(),
      queryFn: getObjectiveStatuses,
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
      queryKey: workspaceKeys.lists(),
      queryFn: () => getWorkspaces(session?.token),
    }),
  ]);

  return (
    <SessionProvider session={session}>
      <HydrationBoundary state={dehydrate(queryClient)}>
        <ApplicationLayout>{children}</ApplicationLayout>
      </HydrationBoundary>
      <OnlineStatusMonitor />
    </SessionProvider>
  );
}
