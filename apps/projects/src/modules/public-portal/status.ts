import type { PublicRequestStatus } from "./types";

export const requestStatusMeta: Record<
  PublicRequestStatus,
  {
    label: string;
    dotClassName: string;
    badgeClassName: string;
  }
> = {
  pending: {
    label: "Pending",
    dotClassName: "bg-warning",
    badgeClassName: "border-warning/30 text-warning bg-warning/10",
  },
  reviewing: {
    label: "Reviewing",
    dotClassName: "bg-info",
    badgeClassName: "border-info/30 text-info bg-info/10",
  },
  planned: {
    label: "Planned",
    dotClassName: "bg-primary",
    badgeClassName: "border-primary/30 text-primary bg-primary/10",
  },
  in_progress: {
    label: "In Progress",
    dotClassName: "bg-secondary",
    badgeClassName: "border-secondary/30 text-secondary bg-secondary/10",
  },
  completed: {
    label: "Completed",
    dotClassName: "bg-success",
    badgeClassName: "border-success/30 text-success bg-success/10",
  },
  closed: {
    label: "Closed",
    dotClassName: "bg-text-muted",
    badgeClassName: "border-border text-text-muted bg-surface-muted/40",
  },
};

export const requestFilters: PublicRequestStatus[] = [
  "pending",
  "reviewing",
  "planned",
  "in_progress",
  "completed",
  "closed",
];

export const roadmapStatuses = [
  "planned",
  "in_progress",
  "completed",
] as const;
