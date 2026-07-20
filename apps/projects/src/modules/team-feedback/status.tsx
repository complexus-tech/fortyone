import { cn } from "lib";
import { Flex, Text } from "ui";
import { Dot } from "@/components/ui";
import type { TeamFeedbackStatus } from "./types";

const statusMeta: Record<
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
    label: "In progress",
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

export const FeedbackStatus = ({ status }: { status: TeamFeedbackStatus }) => {
  const meta = statusMeta[status];

  return (
    <Flex align="center" className="shrink-0" gap={2}>
      <Dot className={cn("size-3", meta.colorClassName)} />
      <Text as="span">{meta.label}</Text>
    </Flex>
  );
};
