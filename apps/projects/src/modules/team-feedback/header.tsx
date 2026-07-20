"use client";

import { cn } from "lib";
import { Button, Flex, Menu, Text } from "ui";
import { CheckIcon, MoreVerticalIcon, RequestsIcon } from "icons";
import { MobileMenuButton } from "@/components/shared";
import { Dot } from "@/components/ui";
import { feedbackStatusMeta } from "./status-meta";
import type { TeamFeedbackListStatus } from "./types";

const filters: {
  colorClassName: string;
  label: string;
  value: TeamFeedbackListStatus;
}[] = [
  {
    colorClassName: "text-foreground",
    label: "All active feedback",
    value: "active",
  },
  {
    colorClassName: "text-text-muted",
    label: "All feedback",
    value: "all",
  },
  {
    colorClassName: feedbackStatusMeta.pending.colorClassName,
    label: "Pending review",
    value: "pending",
  },
  {
    colorClassName: feedbackStatusMeta.reviewing.colorClassName,
    label: "In review",
    value: "reviewing",
  },
  {
    colorClassName: feedbackStatusMeta.planned.colorClassName,
    label: feedbackStatusMeta.planned.label,
    value: "planned",
  },
  {
    colorClassName: feedbackStatusMeta.in_progress.colorClassName,
    label: feedbackStatusMeta.in_progress.label,
    value: "in_progress",
  },
  {
    colorClassName: feedbackStatusMeta.completed.colorClassName,
    label: feedbackStatusMeta.completed.label,
    value: "completed",
  },
  {
    colorClassName: feedbackStatusMeta.closed.colorClassName,
    label: feedbackStatusMeta.closed.label,
    value: "closed",
  },
];

export const TeamFeedbackHeader = ({
  onStatusChange,
  status,
}: {
  onStatusChange: (status: TeamFeedbackListStatus) => void;
  status: TeamFeedbackListStatus;
}) => (
  <Flex
    align="center"
    className="border-border/60 h-16 border-b-[0.5px] px-4"
    justify="between"
  >
    <Flex align="center" className="gap-2">
      <MobileMenuButton />
      <RequestsIcon className="h-5 w-auto" />
      <Text>Feedback</Text>
    </Flex>
    <Menu>
      <Menu.Button>
        <Button
          aria-label="Filter feedback"
          asIcon
          color="tertiary"
          rightIcon={<MoreVerticalIcon />}
          size="sm"
        />
      </Menu.Button>
      <Menu.Items align="end">
        <Menu.Group className="mt-1 mb-3 px-4">
          <Text color="muted" textOverflow="truncate">
            Filter feedback
          </Text>
        </Menu.Group>
        <Menu.Separator className="mb-1.5" />
        <Menu.Group>
          {filters.map((filter) => (
            <Menu.Item
              key={filter.value}
              onSelect={() => {
                onStatusChange(filter.value);
              }}
            >
              <span className="flex size-5 items-center justify-center">
                {status === filter.value ? (
                  <CheckIcon className="h-4 w-auto" strokeWidth={2.2} />
                ) : null}
              </span>
              <Dot className={cn("size-3", filter.colorClassName)} />
              {filter.label}
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.Items>
    </Menu>
  </Flex>
);
