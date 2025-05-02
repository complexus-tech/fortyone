import type { Metadata } from "next";
import type { ReactNode } from "react";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { Box } from "ui";
import { BodyContainer } from "@/components/shared";
import { ListNotifications } from "@/modules/notifications/list";
import { getQueryClient } from "@/app/get-query-client";
import { getNotifications } from "@/modules/notifications/queries/get-notifications";
import { notificationKeys } from "@/constants/keys";
import { auth } from "@/auth";

export const metadata: Metadata = {
  title: "Notifications",
};

export default async function Layout({ children }: { children: ReactNode }) {
  const queryClient = getQueryClient();
  const session = await auth();
  queryClient.prefetchQuery({
    queryKey: notificationKeys.all,
    queryFn: () => getNotifications(session!),
  });
  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <BodyContainer className="grid h-dvh md:grid-cols-[320px_auto]">
        <ListNotifications />
        <Box className="hidden md:block">{children}</Box>
      </BodyContainer>
    </HydrationBoundary>
  );
}
