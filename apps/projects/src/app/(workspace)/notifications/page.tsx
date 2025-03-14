import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { BodyContainer } from "@/components/shared";
import { ListNotifications } from "@/modules/notifications/list";
import { getQueryClient } from "@/app/get-query-client";
import { notificationKeys } from "@/constants/keys";
import { getNotifications } from "@/modules/notifications/queries/get-notifications";
import { storyKeys } from "@/modules/stories/constants";
import { getStory } from "@/modules/story/queries/get-story";

export const metadata: Metadata = {
  title: "Notifications",
};

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ entityId: string; entityType: string }>;
}) {
  const queryClient = getQueryClient();
  const { entityId, entityType } = await searchParams;

  await queryClient.prefetchQuery({
    queryKey: notificationKeys.all,
    queryFn: getNotifications,
  });

  if (entityId && entityType) {
    if (entityType === "story") {
      await queryClient.prefetchQuery({
        queryKey: storyKeys.detail(entityId),
        queryFn: () => getStory(entityId),
      });
    }
  }

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <BodyContainer className="h-screen">
        <ListNotifications />
      </BodyContainer>
    </HydrationBoundary>
  );
}
