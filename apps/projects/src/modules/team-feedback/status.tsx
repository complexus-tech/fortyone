import { cn } from "lib";
import type { TeamFeedbackStatus } from "./types";

const statusMeta: Record<
  TeamFeedbackStatus,
  { label: string; className: string; dotClassName: string }
> = {
  pending: {
    label: "Pending",
    className: "border-border bg-surface text-text-muted",
    dotClassName: "bg-text-muted",
  },
  reviewing: {
    label: "Reviewing",
    className: "border-info/20 bg-info/10 text-info",
    dotClassName: "bg-info",
  },
  planned: {
    label: "Planned",
    className: "border-primary/20 bg-primary/10 text-primary",
    dotClassName: "bg-primary",
  },
  in_progress: {
    label: "In progress",
    className: "border-warning/20 bg-warning/10 text-warning",
    dotClassName: "bg-warning",
  },
  completed: {
    label: "Completed",
    className: "border-success/20 bg-success/10 text-success",
    dotClassName: "bg-success",
  },
  closed: {
    label: "Closed",
    className: "border-danger/20 bg-danger/10 text-danger",
    dotClassName: "bg-danger",
  },
};

export const feedbackStatusLabel = (status: TeamFeedbackStatus) =>
  statusMeta[status].label;

export const FeedbackStatusPill = ({
  compact = false,
  status,
}: {
  compact?: boolean;
  status: TeamFeedbackStatus;
}) => {
  const meta = statusMeta[status];

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-lg border font-medium",
        compact ? "h-6 px-2 text-[0.82rem]" : "h-7 px-2.5 text-[0.92rem]",
        meta.className,
      )}
    >
      <span
        className={cn(
          "rounded-full",
          compact ? "size-1.5" : "size-2",
          meta.dotClassName,
        )}
      />
      {meta.label}
    </span>
  );
};
