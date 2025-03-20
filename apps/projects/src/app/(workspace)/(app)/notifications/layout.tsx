import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { Box } from "ui";
import type { ReactNode } from "react";
import { BodyContainer } from "@/components/shared";
import { NotificationsContainer } from "@/modules/notifications/container";
import { getQueryClient } from "@/app/get-query-client";
import { notificationKeys } from "@/constants/keys";
import { getNotifications } from "@/modules/notifications/queries/get-notifications";

export const metadata: Metadata = {
  title: "Notifications",
};

export default function Layout({ children }: { children: ReactNode }) {
  const queryClient = getQueryClient();

  queryClient.prefetchQuery({
    queryKey: notificationKeys.all,
    queryFn: getNotifications,
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <BodyContainer className="h-screen">
        <Box className="grid grid-cols-[320px_auto]">
          <NotificationsContainer />
          {children}
        </Box>
      </BodyContainer>
    </HydrationBoundary>
  );
}
