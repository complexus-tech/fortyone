"use client";
import { Box, Button, Divider, Flex, Popover, Text } from "ui";
import { ArrowDownIcon, AvatarIcon, CheckIcon, FilterIcon } from "icons";
import { useRef, type ReactNode } from "react";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";

export type StoriesFilter = {
  statusIds: string[] | null;
  assigneeIds: string[] | null;
  reporterIds: string[] | null;
  priorities: string[] | null;
  teamIds: string[] | null;
  sprintIds: string[] | null;
  labelIds: string[] | null;
  parentId: string | null;
  objectiveId: string | null;
  epicId: string | null;
  keyResultId: string | null;
  hasNoAssignee: boolean | null;
  assignedToMe: boolean;
  createdByMe: boolean;
};

type StoriesFilterButtonProps = {
  filters: StoriesFilter;
  setFilters: (v: StoriesFilter) => void;
  resetFilters: () => void;
};

const ToggleButton = ({
  label,
  icon,
  isActive,
  onClick,
}: {
  label: string;
  isActive: boolean;
  icon: ReactNode;
  onClick: () => void;
}) => {
  return (
    <button
      className={cn(
        "flex w-full items-center justify-between px-4 py-3 transition hover:bg-gray-50 hover:dark:bg-dark-200",
      )}
      onClick={onClick}
      type="button"
    >
      <span className="flex items-center gap-3">
        {icon}
        {label}
      </span>
      {isActive ? <CheckIcon className="h-5 w-auto text-primary" /> : null}
    </button>
  );
};

export const StoriesFilterButton = ({
  filters,
  setFilters,
  resetFilters,
}: StoriesFilterButtonProps) => {
  const buttonRef = useRef<HTMLButtonElement>(null);

  // filtersCount returns the number of filters applied.
  const filtersCount = () => {
    let count = 0;

    // Count array filters
    const arrayFilters = [
      "statusIds",
      "assigneeIds",
      "reporterIds",
      "priorities",
      "teamIds",
      "sprintIds",
      "labelIds",
    ] as const;
    arrayFilters.forEach((key) => {
      if (filters[key] && filters[key].length > 0) count++;
    });

    // Count string filters
    const stringFilters = [
      "parentId",
      "objectiveId",
      "epicId",
      "keyResultId",
    ] as const;
    stringFilters.forEach((key) => {
      if (filters[key]) count++;
    });

    // Count boolean filters
    if (filters.hasNoAssignee) count++;
    if (filters.assignedToMe) count++;
    if (filters.createdByMe) count++;

    return count;
  };

  const getButtonLabel = () => {
    if (filtersCount()) {
      return `${filtersCount()} filter${filtersCount() > 1 ? "s" : ""} applied`;
    }
    return "Filters";
  };

  useHotkeys("v+f", (e) => {
    e.preventDefault();
    buttonRef.current?.click();
  });

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          color="tertiary"
          leftIcon={<FilterIcon className="h-4 w-auto" />}
          ref={buttonRef}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          <span className="hidden md:inline">{getButtonLabel()}</span>
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-[26rem] pb-5">
        <Flex align="center" className="h-10 px-4" justify="between">
          <Text
            color="muted"
            fontSize="sm"
            fontWeight="semibold"
            transform="uppercase"
          >
            Apply Filters
          </Text>
          {filtersCount() > 0 && (
            <Button
              className="text-primary dark:text-primary"
              color="tertiary"
              onClick={resetFilters}
              size="sm"
              variant="naked"
            >
              Clear filters
            </Button>
          )}
        </Flex>
        <Divider className="mt-1.5" />
        <Box>
          <ToggleButton
            icon={<AvatarIcon className="h-5 w-auto" />}
            isActive={filters.assignedToMe}
            label="Assigned to me"
            onClick={() => {
              setFilters({ ...filters, assignedToMe: !filters.assignedToMe });
            }}
          />
          <ToggleButton
            icon={<AvatarIcon className="h-5 w-auto" />}
            isActive={filters.createdByMe}
            label="Created by me"
            onClick={() => {
              setFilters({ ...filters, createdByMe: !filters.createdByMe });
            }}
          />
          <ToggleButton
            icon={<AvatarIcon className="h-5 w-auto" />}
            isActive={filters.hasNoAssignee || false}
            label="Has no assignee"
            onClick={() => {
              setFilters({ ...filters, hasNoAssignee: !filters.hasNoAssignee });
            }}
          />
        </Box>
      </Popover.Content>
    </Popover>
  );
};
