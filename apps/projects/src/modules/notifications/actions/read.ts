import { getNotificationReadPath } from "lib/src/notification-read";
import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const readNotification = async (
  notificationId: string,
  workspaceSlug: string,
  observedCreatedAt?: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    await put(
      getNotificationReadPath(notificationId, observedCreatedAt),
      {},
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
