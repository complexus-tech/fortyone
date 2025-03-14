import { redirect } from "next/navigation";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { getStory } from "@/modules/story/queries/get-story";
import { storyKeys } from "@/modules/stories/constants";
import { NotificationDetails } from "@/modules/notifications/details";

export default async function Page({
  params,
  searchParams,
}: {
  params: Promise<{ notificationId: string }>;
  searchParams: Promise<{
    entityId?: string;
    entityType?: "story" | "objective";
  }>;
}) {
  const { notificationId } = await params;
  const { entityId, entityType } = await searchParams;
  const queryClient = getQueryClient();

  if (!entityId || !entityType) {
    return redirect("/notifications");
  }

  if (entityType === "story") {
    await queryClient.prefetchQuery({
      queryKey: storyKeys.detail(entityId),
      queryFn: () => getStory(entityId),
    });
  }

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <NotificationDetails
        entityId={entityId}
        entityType={entityType}
        notificationId={notificationId}
      />
    </HydrationBoundary>
  );
}
