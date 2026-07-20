import { cn } from "lib";
import { Flex, Text } from "ui";
import { Dot } from "@/components/ui";
import { feedbackStatusMeta } from "./status-meta";
import type { TeamFeedbackStatus } from "./types";

export const FeedbackStatus = ({ status }: { status: TeamFeedbackStatus }) => {
  const meta = feedbackStatusMeta[status];

  return (
    <Flex align="center" className="shrink-0" gap={2}>
      <Dot className={cn("size-3", meta.colorClassName)} />
      <Text as="span">{meta.label}</Text>
    </Flex>
  );
};
