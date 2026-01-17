import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { NotificationDetails } from "@/modules/notifications/details";
import { getStory } from "@/modules/story/queries/get-story";
import { auth } from "@/auth";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { withWorkspacePath } from "@/utils";

export async function generateMetadata({
  params,
  searchParams,
}: {
  params: Promise<{ notificationId: string; workspaceSlug: string }>;
  searchParams: Promise<{
    entityId: string;
    entityType?: "story" | "objective";
  }>;
}): Promise<Metadata> {
  const { workspaceSlug } = await params;
  const { entityType, entityId } = await searchParams;
  const session = await auth();
  const ctx = { session: session!, workspaceSlug };

  let title = "Notification Details";
  if (entityType === "story") {
    const storyData = await getStory(entityId, ctx);

    title = storyData?.title || "Story";
  } else if (entityType === "objective") {
    const objectiveData = await getObjective(entityId, ctx);
    title = objectiveData?.name || "Objective";
  }
  return {
    title,
  };
}

export default async function Page({
  params,
  searchParams,
}: {
  params: Promise<{ notificationId: string; workspaceSlug: string }>;
  searchParams: Promise<{
    entityId?: string;
    entityType?: "story" | "objective";
  }>;
}) {
  const { notificationId, workspaceSlug } = await params;
  const { entityId, entityType } = await searchParams;

  if (!entityId || !entityType) {
    return redirect(withWorkspacePath("/notifications", workspaceSlug));
  }

  return (
    <NotificationDetails
      entityId={entityId}
      entityType={entityType}
      notificationId={notificationId}
    />
  );
}
