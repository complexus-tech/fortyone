import type { Metadata } from "next";
import type { ReactNode } from "react";
import { SessionProvider } from "next-auth/react";
import { SettingsLayout } from "@/components/layouts";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { labelKeys, memberKeys, teamKeys } from "@/constants/keys";
import { getMembers } from "@/lib/queries/members/get-members";
import { getTeams } from "@/modules/teams/queries/get-teams";

export const metadata: Metadata = {
  title: "Settings",
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();
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
  ]);

  return (
    <SessionProvider session={session}>
      <SettingsLayout>{children}</SettingsLayout>
    </SessionProvider>
  );
}
