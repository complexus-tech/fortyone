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
import { memberKeys , labelKeys } from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/contants";
import { getLabels } from "@/lib/queries/labels/get-labels";

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const queryClient = getQueryClient();
  const session = await auth();

  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: ["statuses"],
      queryFn: getStatuses,
    }),
    queryClient.prefetchQuery({
      queryKey: ["teams"],
      queryFn: getTeams,
    }),
    queryClient.prefetchQuery({
      queryKey: ["sprints"],
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
  ]);

  return (
    <SessionProvider session={session}>
      <HydrationBoundary state={dehydrate(queryClient)}>
        <ApplicationLayout>{children}</ApplicationLayout>
      </HydrationBoundary>
    </SessionProvider>
  );
}
