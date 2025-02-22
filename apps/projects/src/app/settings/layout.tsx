import type { Metadata } from "next";
import type { ReactNode } from "react";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { SettingsLayout } from "@/components/layouts";
import { getQueryClient } from "@/app/get-query-client";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { labelKeys, memberKeys, teamKeys, userKeys } from "@/constants/keys";
import { getMembers } from "@/lib/queries/members/get-members";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getProfile } from "@/lib/queries/users/profile";

export const metadata: Metadata = {
  title: "Settings",
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const queryClient = getQueryClient();

  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: labelKeys.lists(),
      queryFn: () => getLabels(),
    }),
    queryClient.prefetchQuery({
      queryKey: memberKeys.lists(),
      queryFn: () => getMembers(),
    }),
    queryClient.prefetchQuery({
      queryKey: teamKeys.lists(),
      queryFn: () => getTeams(),
    }),
    queryClient.prefetchQuery({
      queryKey: userKeys.profile(),
      queryFn: () => getProfile(),
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <SettingsLayout>{children}</SettingsLayout>
    </HydrationBoundary>
  );
}
