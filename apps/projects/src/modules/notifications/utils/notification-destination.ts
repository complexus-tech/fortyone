import type { AppNotification } from "../types";

export type NotificationEntityType = AppNotification["entityType"];

const NOTIFICATION_ENTITY_TYPES = new Set<NotificationEntityType>([
  "story",
  "objective",
  "key_result",
  "strategy",
]);

const COMMENT_ID_PATTERN =
  /^[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}$/i;

export const getNotificationCommentId = (
  message: AppNotification["message"] | undefined,
) => {
  const value = message?.variables.commentId?.value;
  return value &&
    COMMENT_ID_PATTERN.test(value) &&
    value !== "00000000-0000-0000-0000-000000000000"
    ? value
    : undefined;
};

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
  observedCreatedAt,
  message,
}: Pick<AppNotification, "entityId" | "entityType"> & {
  notificationId: string;
  observedCreatedAt?: string;
  message?: AppNotification["message"];
}) => {
  const searchParams = new URLSearchParams({ entityId, entityType });
  if (observedCreatedAt)
    searchParams.set("observedCreatedAt", observedCreatedAt);
  const commentId = getNotificationCommentId(message);
  if (entityType === "story" && commentId)
    searchParams.set("comment", commentId);

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
