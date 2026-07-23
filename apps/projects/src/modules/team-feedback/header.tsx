"use client";

import { Button, Flex, Menu, Text } from "ui";
import { CheckIcon, MoreVerticalIcon, RequestsIcon } from "icons";
import { ExpandableSearchHeader, MobileMenuButton } from "@/components/shared";
import { useUserRole } from "@/hooks/role";
import type { TeamFeedbackListStatus } from "./types";

type FeedbackFilter = {
  label: string;
  value: TeamFeedbackListStatus;
};

const activeFilter: FeedbackFilter = {
  label: "Active",
  value: "active",
};

const ongoingStatusFilters: FeedbackFilter[] = [
  {
    label: "Pending",
    value: "pending",
  },
  {
    label: "In Review",
    value: "reviewing",
  },
  {
    label: "Planned",
    value: "planned",
  },
  {
    label: "In Progress",
    value: "in_progress",
  },
];

const terminalStatusFilters: FeedbackFilter[] = [
  {
    label: "Completed",
    value: "completed",
  },
  {
    label: "Closed",
    value: "closed",
  },
];

const trashFilter: FeedbackFilter = {
  label: "Trash",
  value: "trashed",
};

const FeedbackFilterItem = ({
  filter,
  isSelected,
  onSelect,
}: {
  filter: FeedbackFilter;
  isSelected: boolean;
  onSelect: () => void;
}) => (
  <Menu.Item onSelect={onSelect}>
    <span className="flex size-5 items-center justify-center">
      {isSelected ? (
        <CheckIcon className="h-4 w-auto" strokeWidth={2.2} />
      ) : null}
    </span>
    {filter.label}
  </Menu.Item>
);

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
              <FeedbackFilterItem
                filter={activeFilter}
                isSelected={status === activeFilter.value}
                onSelect={() => {
                  onStatusChange(activeFilter.value);
                }}
              />
            </Menu.Group>
            <Menu.Separator className="my-1.5" />
            <Menu.Group>
              {ongoingStatusFilters.map((filter) => (
                <FeedbackFilterItem
                  filter={filter}
                  isSelected={status === filter.value}
                  key={filter.value}
                  onSelect={() => {
                    onStatusChange(filter.value);
                  }}
                />
              ))}
            </Menu.Group>
            <Menu.Separator className="my-1.5" />
            <Menu.Group>
              {terminalStatusFilters.map((filter) => (
                <FeedbackFilterItem
                  filter={filter}
                  isSelected={status === filter.value}
                  key={filter.value}
                  onSelect={() => {
                    onStatusChange(filter.value);
                  }}
                />
              ))}
            </Menu.Group>
            {userRole === "admin" ? (
              <>
                <Menu.Separator className="my-1.5" />
                <Menu.Group>
                  <FeedbackFilterItem
                    filter={trashFilter}
                    isSelected={status === trashFilter.value}
                    onSelect={() => {
                      onStatusChange(trashFilter.value);
                    }}
                  />
                </Menu.Group>
              </>
            ) : null}
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
