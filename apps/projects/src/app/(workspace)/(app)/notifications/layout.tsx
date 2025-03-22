import type { Metadata } from "next";
import type { ReactNode } from "react";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { BodyContainer } from "@/components/shared";
import { ListNotifications } from "@/modules/notifications/list";
import { getQueryClient } from "@/app/get-query-client";
import { getNotifications } from "@/modules/notifications/queries/get-notifications";
import { notificationKeys } from "@/constants/keys";

export const metadata: Metadata = {
  title: "Notifications",
};

export default function Layout({ children }: { children: ReactNode }) {
  const queryClient = getQueryClient();
  queryClient.prefetchQuery({
    queryKey: notificationKeys.all,
    queryFn: () => getNotifications(),
  });
  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <BodyContainer className="grid h-screen grid-cols-[320px_auto]">
        <ListNotifications />
        {children}
      </BodyContainer>
    </HydrationBoundary>
  );
}
