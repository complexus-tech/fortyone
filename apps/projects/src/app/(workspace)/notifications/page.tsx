import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { BodyContainer } from "@/components/shared";
import { ListNotifications } from "@/modules/notifications/list";
import { getQueryClient } from "@/app/get-query-client";
import { notificationKeys } from "@/constants/keys";
import { getNotifications } from "@/modules/notifications/queries/get-notifications";

export const metadata: Metadata = {
  title: "Notifications",
};

export default async function Page() {
  const queryClient = getQueryClient();

  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: notificationKeys.all,
      queryFn: getNotifications,
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <BodyContainer className="h-screen">
        <ListNotifications />
      </BodyContainer>
    </HydrationBoundary>
  );
}
