import { ApiError } from "api-client";

export type NotificationTargetResolution<T> =
  | { status: "found"; value: T }
  | { status: "terminal"; reason: "forbidden" | "not_found" };

/**
 * Resolves a notification target without disguising an operational failure as
 * a deleted or inaccessible entity. Only explicit 403/404 responses (and an
 * empty successful response) are terminal; all other failures remain visible
 * to the route error boundary and leave the notification unread.
 */
export const resolveNotificationTarget = async <T>(
  load: () => Promise<T | null | undefined>,
): Promise<NotificationTargetResolution<T>> => {
  try {
    const value = await load();

    if (value === null || value === undefined) {
      return { reason: "not_found", status: "terminal" };
    }

    return { status: "found", value };
  } catch (error) {
    if (error instanceof ApiError && error.status === 403) {
      return { reason: "forbidden", status: "terminal" };
    }
    if (error instanceof ApiError && error.status === 404) {
      return { reason: "not_found", status: "terminal" };
    }

    throw error;
  }
};
