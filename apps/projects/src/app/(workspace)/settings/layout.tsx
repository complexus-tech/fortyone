import type { Metadata } from "next";
import type { ReactNode } from "react";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { SettingsLayout } from "@/components/layouts";
import { getQueryClient } from "@/app/get-query-client";
import { invitationKeys, notificationKeys } from "@/constants/keys";
import { getPendingInvitations } from "@/modules/invitations/queries/pending-invitations";
import { getNotificationPreferences } from "@/modules/notifications/queries/get-preferences";

export const metadata: Metadata = {
  title: "Settings",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  const queryClient = getQueryClient();

  queryClient.prefetchQuery({
    queryKey: invitationKeys.pending,
    queryFn: getPendingInvitations,
  });
  queryClient.prefetchQuery({
    queryKey: notificationKeys.preferences(),
    queryFn: getNotificationPreferences,
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <SettingsLayout>{children}</SettingsLayout>
    </HydrationBoundary>
  );
}
