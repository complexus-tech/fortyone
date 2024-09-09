import type { Metadata } from "next";
import type { ReactNode } from "react";
import { SessionProvider } from "next-auth/react";
import { ApplicationLayout } from "@/components/layouts";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";

export const metadata: Metadata = {
  title: "Objectives",
  description: "Complexus Objectives",
};

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
      queryKey: ["objectives"],
      queryFn: getObjectives,
    }),
    queryClient.prefetchQuery({
      queryKey: ["teams"],
      queryFn: getTeams,
    }),
    queryClient.prefetchQuery({
      queryKey: ["sprints"],
      queryFn: getSprints,
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
