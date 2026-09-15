import type { AppNotification } from "../types";
import { teamStoriesHref } from "@/modules/teams/stories/team-story-navigation";

export type NotificationDestination =
  | { type: "story"; storyId: string }
  | { type: "teamStories"; href: ReturnType<typeof teamStoriesHref> }
  | { type: "summary" };

export const resolveNotificationDestination = async (
  notification: Pick<AppNotification, "id" | "entityId" | "entityType">,
  context: {
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
    return {
      type: "teamStories",
      href: teamStoriesHref(
        notification.entityType === "objective"
          ? { objectiveId: entity.id, teamId: entity.teamId }
          : { sprintId: entity.id, teamId: entity.teamId },
      ),
    };
  }
  return { type: "summary" };
};
