import type { AppNotification } from "../types";

export type NotificationDestination =
  | { type: "story"; storyId: string }
  | { type: "objective"; objectiveId: string; teamId: string }
  | { type: "sprint"; sprintId: string; teamId: string }
  | { type: "web"; url: string };

export const notificationWebURL = (
  applicationURL: string,
  workspace: string,
  notification: Pick<AppNotification, "id" | "entityId" | "entityType">,
) => {
  const url = new URL(applicationURL);
  if (!/^[a-z0-9][a-z0-9-]*$/.test(workspace))
    throw new Error("The workspace address is invalid.");
  const hosted =
    url.hostname === "fortyone.app" || url.hostname.endsWith(".fortyone.app");
  if (hosted) url.hostname = `${workspace}.fortyone.app`;
  const path =
    notification.entityType === "strategy" ? "/strategy" : "/notifications";
  url.pathname = hosted
    ? path
    : `${url.pathname.replace(/\/$/, "")}/${encodeURIComponent(workspace)}${path}`;
  url.search = "";
  url.hash = "";
  return url.toString();
};

export const resolveNotificationDestination = async (
  notification: Pick<AppNotification, "id" | "entityId" | "entityType">,
  context: {
    applicationURL: string;
    workspace: string;
    loadObjective: (
      id: string,
    ) => Promise<{ id: string; teamId: string } | null | undefined>;
    loadSprint: (
      id: string,
    ) => Promise<{ id: string; teamId: string } | null | undefined>;
  },
): Promise<NotificationDestination> => {
  if (notification.entityType === "story")
    return { type: "story", storyId: notification.entityId };
  if (
    notification.entityType === "objective" ||
    notification.entityType === "sprint"
  ) {
    const load =
      notification.entityType === "objective"
        ? context.loadObjective
        : context.loadSprint;
    const entity = await load(notification.entityId);
    if (!entity || entity.id !== notification.entityId || !entity.teamId) {
      throw new Error("This item is no longer available in your workspace.");
    }
    return notification.entityType === "objective"
      ? { type: "objective", objectiveId: entity.id, teamId: entity.teamId }
      : { type: "sprint", sprintId: entity.id, teamId: entity.teamId };
  }
  return {
    type: "web",
    url: notificationWebURL(
      context.applicationURL,
      context.workspace,
      notification,
    ),
  };
};
