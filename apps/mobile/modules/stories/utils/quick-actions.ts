import type { Status } from "@/types/statuses";
import type { Story } from "../types";

export function getCompletionStatus(statuses: Status[], teamId: string) {
  return statuses
    .filter(
      (status) => status.teamId === teamId && status.category === "completed",
    )
    .sort((a, b) => a.orderIndex - b.orderIndex || a.id.localeCompare(b.id))[0];
}

export function canCompleteStory(
  story: Pick<
    Story,
    "teamId" | "statusId" | "completedAt" | "archivedAt" | "deletedAt"
  >,
  status?: Status,
  completionStatus?: Status,
) {
  return Boolean(
    status &&
      completionStatus &&
      status.id === story.statusId &&
      status.teamId === story.teamId &&
      completionStatus.teamId === story.teamId &&
      completionStatus.category === "completed" &&
      status.category !== "completed" &&
      status.category !== "cancelled" &&
      !story.completedAt &&
      !story.archivedAt &&
      !story.deletedAt,
  );
}
