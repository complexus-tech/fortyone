import type { AppNotification } from "../types";

export type NotificationEntityType = AppNotification["entityType"];

const NOTIFICATION_ENTITY_TYPES = new Set<NotificationEntityType>([
  "story",
  "objective",
  "key_result",
  "strategy",
]);

export const isNotificationEntityType = (
  value: string | undefined,
): value is NotificationEntityType =>
  Boolean(
    value && NOTIFICATION_ENTITY_TYPES.has(value as NotificationEntityType),
  );

export const getSingleSearchParam = (value: string | string[] | undefined) =>
  typeof value === "string" && value.trim() ? value.trim() : undefined;

export const getNotificationDetailsPath = ({
  entityId,
  entityType,
  notificationId,
}: Pick<AppNotification, "entityId" | "entityType"> & {
  notificationId: string;
}) => {
  const searchParams = new URLSearchParams({ entityId, entityType });

  return `/notifications/${encodeURIComponent(notificationId)}?${searchParams.toString()}`;
};

export const getObjectiveDetailsPath = ({
  keyResultId,
  objectiveId,
  teamId,
}: {
  keyResultId?: string;
  objectiveId: string;
  teamId: string;
}) => {
  const searchParams = new URLSearchParams({ tab: "overview" });
  if (keyResultId) searchParams.set("keyResultId", keyResultId);

  return `/teams/${encodeURIComponent(teamId)}/objectives/${encodeURIComponent(objectiveId)}?${searchParams.toString()}`;
};
