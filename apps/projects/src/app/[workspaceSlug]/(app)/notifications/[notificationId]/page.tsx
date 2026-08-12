import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { NotificationDetails } from "@/modules/notifications/details";
import { getStory } from "@/modules/story/queries/get-story";
import { auth } from "@/auth";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { getKeyResults } from "@/modules/objectives/queries/get-key-results";
import { readNotification } from "@/modules/notifications/actions/read";
import {
  getObjectiveDetailsPath,
  getSingleSearchParam,
  isNotificationEntityType,
} from "@/modules/notifications/utils/notification-destination";
import { resolveNotificationTarget } from "@/modules/notifications/utils/resolve-notification-target";
import { withWorkspacePath } from "@/utils";

type NotificationSearchParams = {
  entityId?: string | string[];
  entityType?: string | string[];
  objectiveId?: string | string[];
};

export async function generateMetadata({
  params,
  searchParams,
}: {
  params: Promise<{ notificationId: string; workspaceSlug: string }>;
  searchParams: Promise<NotificationSearchParams>;
}): Promise<Metadata> {
  const [{ workspaceSlug }, rawSearchParams, session] = await Promise.all([
    params,
    searchParams,
    auth(),
  ]);
  const entityId = getSingleSearchParam(rawSearchParams.entityId);
  const entityTypeValue = getSingleSearchParam(rawSearchParams.entityType);
  const entityType = isNotificationEntityType(entityTypeValue)
    ? entityTypeValue
    : undefined;
  const ctx = { session: session!, workspaceSlug };

  let title = "Notification Details";
  if (entityType === "story" && entityId) {
    const storyResolution = await resolveNotificationTarget(() =>
      getStory(entityId, ctx),
    );

    title =
      storyResolution.status === "found"
        ? storyResolution.value.title
        : "Story";
  } else if (entityType === "objective" && entityId) {
    const objectiveResolution = await resolveNotificationTarget(() =>
      getObjective(entityId, ctx),
    );
    title =
      objectiveResolution.status === "found"
        ? objectiveResolution.value.name
        : "Objective";
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
  searchParams: Promise<NotificationSearchParams>;
}) {
  const { notificationId, workspaceSlug } = await params;
  const rawSearchParams = await searchParams;
  const entityId = getSingleSearchParam(rawSearchParams.entityId);
  const entityTypeValue = getSingleSearchParam(rawSearchParams.entityType);
  const objectiveId = getSingleSearchParam(rawSearchParams.objectiveId);
  const notificationsPath = withWorkspacePath("/notifications", workspaceSlug);

  if (!entityId || !isNotificationEntityType(entityTypeValue)) {
    await readNotification(notificationId, workspaceSlug);
    return redirect(notificationsPath);
  }

  const entityType = entityTypeValue;
  if (entityType === "strategy") {
    await readNotification(notificationId, workspaceSlug);
    return redirect(withWorkspacePath("/strategy", workspaceSlug));
  }

  if (entityType === "objective" || entityType === "key_result") {
    const parentObjectiveId =
      entityType === "key_result" ? objectiveId : entityId;
    if (!parentObjectiveId) {
      await readNotification(notificationId, workspaceSlug);
      return redirect(notificationsPath);
    }

    const session = await auth();
    const ctx = {
      session: session!,
      workspaceSlug,
    };
    const objectivePromise = resolveNotificationTarget(() =>
      getObjective(parentObjectiveId, ctx),
    );
    const keyResultPromise =
      entityType === "key_result"
        ? resolveNotificationTarget(async () => {
            const keyResults = await getKeyResults(parentObjectiveId, ctx);
            return keyResults.find((keyResult) => keyResult.id === entityId);
          })
        : undefined;
    const [objectiveResolution, keyResultResolution] = await Promise.all([
      objectivePromise,
      keyResultPromise,
    ]);

    if (
      objectiveResolution.status === "terminal" ||
      keyResultResolution?.status === "terminal" ||
      objectiveResolution.value.id !== parentObjectiveId
    ) {
      await readNotification(notificationId, workspaceSlug);
      return redirect(notificationsPath);
    }

    await readNotification(notificationId, workspaceSlug);
    return redirect(
      withWorkspacePath(
        getObjectiveDetailsPath({
          keyResultId: entityType === "key_result" ? entityId : undefined,
          objectiveId: objectiveResolution.value.id,
          teamId: objectiveResolution.value.teamId,
        }),
        workspaceSlug,
      ),
    );
  }

  return (
    <NotificationDetails entityId={entityId} notificationId={notificationId} />
  );
}
