"use client";

import { cn } from "lib";
import { Button, Flex, Menu, Text } from "ui";
import { CheckIcon, DeleteIcon, MoreVerticalIcon, RequestsIcon } from "icons";
import { ExpandableSearchHeader, MobileMenuButton } from "@/components/shared";
import { Dot } from "@/components/ui";
import { useUserRole } from "@/hooks/role";
import { feedbackStatusMeta } from "./status-meta";
import type { TeamFeedbackListStatus } from "./types";

const filters: {
  colorClassName: string;
  label: string;
  value: TeamFeedbackListStatus;
}[] = [
  {
    colorClassName: "text-foreground",
    label: "Active Feedback",
    value: "active",
  },
  {
    colorClassName: feedbackStatusMeta.pending.colorClassName,
    label: "Pending Review",
    value: "pending",
  },
  {
    colorClassName: feedbackStatusMeta.reviewing.colorClassName,
    label: "In Review",
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
  {
    colorClassName: "text-danger",
    label: "Trash",
    value: "trashed",
  },
];

export const TeamFeedbackHeader = ({
  onSearchChange,
  onStatusChange,
  search,
  status,
}: {
  onSearchChange: (search: string) => void;
  onStatusChange: (status: TeamFeedbackListStatus) => void;
  search: string;
  status: TeamFeedbackListStatus;
}) => {
  const { userRole } = useUserRole();
  const visibleFilters =
    userRole === "admin"
      ? filters
      : filters.filter((filter) => filter.value !== "trashed");

  return (
    <ExpandableSearchHeader
      actions={
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
              {visibleFilters.map((filter) => (
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
                  {filter.value === "trashed" ? (
                    <DeleteIcon
                      className={cn("size-4", filter.colorClassName)}
                    />
                  ) : (
                    <Dot className={cn("size-3", filter.colorClassName)} />
                  )}
                  {filter.label}
                </Menu.Item>
              ))}
            </Menu.Group>
          </Menu.Items>
        </Menu>
      }
      initialValue={search}
      key={search}
      label="Search feedback"
      leading={
        <Flex align="center" className="gap-2">
          <MobileMenuButton />
          <RequestsIcon className="h-5 w-auto" />
          <Text>Feedback</Text>
        </Flex>
      }
      onSubmit={onSearchChange}
      placeholder="Search feedback..."
    />
  );
};
