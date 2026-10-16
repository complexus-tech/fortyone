import { renderNotificationTemplate } from "lib/src/notification-template";
import type { DesktopNotification, DesktopNotificationScope } from "./bridge";
import { getDesktopBridge } from "./bridge";
import { bindDesktopSession } from "./session-binding";

const UUID = /^[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}$/i;
const TIMESTAMP =
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/;
const ENTITY_TYPES = new Set([
  "story",
  "comment",
  "objective",
  "key_result",
  "strategy",
]);
const isRecord = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === "object" && !Array.isArray(value);

// The SSE domain payload uses snake case; API-shaped events use camel case.
// Reject contradictory aliases instead of choosing a potentially wrong scope.
const field = (value: Record<string, unknown>, camel: string, snake: string) =>
  camel in value && snake in value && value[camel] !== value[snake]
    ? undefined
    : value[camel] ?? value[snake];

/** Only live, unread notifications for this authenticated subscription qualify. */
export const parseDesktopNotification = (
  serialized: unknown,
  scope: DesktopNotificationScope,
): DesktopNotification | null => {
  if (typeof serialized !== "string" || serialized.length > 128 * 1024)
    return null;
  let value: unknown;
  try {
    value = JSON.parse(serialized) as unknown;
  } catch {
    return null;
  }
  if (!isRecord(value)) return null;
  const recipientId = field(value, "recipientId", "recipient_id");
  const workspaceId = field(value, "workspaceId", "workspace_id");
  const entityId = field(value, "entityId", "entity_id");
  const entityType = field(value, "entityType", "entity_type");
  const createdAt = field(value, "createdAt", "created_at");
  if (
    recipientId !== scope.accountId ||
    workspaceId !== scope.workspaceId ||
    ![value.id, recipientId, workspaceId, entityId].every(
      (id) => typeof id === "string" && UUID.test(id),
    ) ||
    typeof entityType !== "string" ||
    !ENTITY_TYPES.has(entityType) ||
    typeof createdAt !== "string" ||
    !TIMESTAMP.test(createdAt) ||
    !Number.isFinite(Date.parse(createdAt)) ||
    typeof value.title !== "string" ||
    !value.title.trim() ||
    value.title.length > 2048 ||
    (value.readAt !== null && value.readAt !== undefined) ||
    (value.read_at !== null && value.read_at !== undefined) ||
    !isRecord(value.message) ||
    typeof value.message.template !== "string" ||
    value.message.template.length > 4096
  )
    return null;

  const variables = value.message.variables;
  if (
    variables !== null &&
    variables !== undefined &&
    (!isRecord(variables) ||
      Object.keys(variables).length > 64 ||
      Object.values(variables).some(
        (variable) =>
          !isRecord(variable) ||
          typeof variable.value !== "string" ||
          variable.value.length > 4096 ||
          (variable.type !== undefined && typeof variable.type !== "string"),
      ))
  )
    return null;

  const { text } = renderNotificationTemplate({
    template: value.message.template,
    variables: variables as Record<
      string,
      { value: string; type?: string }
    > | null,
  });
  const commentVariable = isRecord(variables) ? variables.commentId : null;
  const commentId = isRecord(commentVariable) ? commentVariable.value : null;
  return {
    ...scope,
    notificationId: value.id as string,
    entityId: entityId as string,
    entityType,
    createdAt,
    title: value.title.trim().slice(0, 256),
    body: text.slice(0, 1024),
    ...(entityType === "story" &&
    typeof commentId === "string" &&
    UUID.test(commentId) &&
    commentId !== "00000000-0000-0000-0000-000000000000"
      ? { commentId }
      : {}),
  };
};

export const deliverDesktopNotification = async (
  serialized: unknown,
  scope: DesktopNotificationScope,
  isCurrent: () => boolean,
) => {
  const bridge = getDesktopBridge();
  if (!bridge?.deliverNotification || !isCurrent()) return;
  const notification = parseDesktopNotification(serialized, scope);
  if (!notification) return;
  if ((await bindDesktopSession(scope)) && isCurrent()) {
    // Older native builds reject unknown fields in the notification payload.
    const { commentId, ...legacyNotification } = notification;
    await bridge.deliverNotification(
      bridge.commentNotificationLinks && commentId
        ? notification
        : legacyNotification,
    );
  }
};
