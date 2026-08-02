import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { NotificationDetails } from "@/modules/notifications/details";
import { getStory } from "@/modules/story/queries/get-story";
import { auth } from "@/auth";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { readNotification } from "@/modules/notifications/actions/read";
import { withWorkspacePath } from "@/utils";

export async function generateMetadata({
  params,
  searchParams,
}: {
  params: Promise<{ notificationId: string; workspaceSlug: string }>;
  searchParams: Promise<{
    entityId: string;
    entityType?: "story" | "objective" | "key_result" | "strategy";
  }>;
}): Promise<Metadata> {
  const [{ workspaceSlug }, { entityType, entityId }, session] =
    await Promise.all([params, searchParams, auth()]);
  const ctx = { session: session!, workspaceSlug };

  let title = "Notification Details";
  if (entityType === "story") {
    const storyData = await getStory(entityId, ctx);

    title = storyData?.title || "Story";
  } else if (entityType === "objective") {
    const objectiveData = await getObjective(entityId, ctx);
    title = objectiveData?.name || "Objective";
  } else if (entityType === "key_result") {
    title = "Key result update";
  } else if (entityType === "strategy") {
    title = "Strategy update";
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
    entityType?: "story" | "objective" | "key_result" | "strategy";
  }>;
}) {
  const { notificationId, workspaceSlug } = await params;
  const { entityId, entityType } = await searchParams;

  if (!entityId || !entityType) {
    return redirect(withWorkspacePath("/notifications", workspaceSlug));
  }

  if (entityType === "strategy") {
    await readNotification(notificationId, workspaceSlug);
    return redirect(withWorkspacePath("/strategy", workspaceSlug));
  }

  if (entityType === "objective" || entityType === "key_result") {
    const session = await auth();
    const objective = await getObjective(entityId, {
      session: session!,
      workspaceSlug,
    });
    await readNotification(notificationId, workspaceSlug);
    if (!objective) {
      return redirect(withWorkspacePath("/notifications", workspaceSlug));
    }
    return redirect(
      withWorkspacePath(
        `/teams/${objective.teamId}/objectives/${objective.id}`,
        workspaceSlug,
      ),
    );
  }

  return (
    <NotificationDetails entityId={entityId} notificationId={notificationId} />
  );
}
