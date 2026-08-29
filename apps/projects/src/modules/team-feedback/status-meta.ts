import type { TeamFeedbackStatus } from "./types";

export const feedbackStatusMeta: Record<
  TeamFeedbackStatus,
  { label: string; colorClassName: string }
> = {
  pending: {
    label: "Pending",
    colorClassName: "text-text-muted",
  },
  reviewing: {
    label: "Reviewing",
    colorClassName: "text-info",
  },
  planned: {
    label: "Planned",
    colorClassName: "text-primary",
  },
  in_progress: {
    label: "In Progress",
    colorClassName: "text-warning",
  },
  completed: {
    label: "Completed",
    colorClassName: "text-success",
  },
  closed: {
    label: "Closed",
    colorClassName: "text-danger",
  },
};
