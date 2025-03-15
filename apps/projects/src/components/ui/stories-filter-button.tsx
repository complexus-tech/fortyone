"use client";
import {
  Avatar,
  Box,
  Button,
  DatePicker,
  Divider,
  Flex,
  Popover,
  Text,
} from "ui";
import {
  ArrowDownIcon,
  ArrowRightIcon,
  AvatarIcon,
  CalendarIcon,
  CheckIcon,
  FilterIcon,
  SprintsIcon,
} from "icons";
import { useState, type ReactNode } from "react";
import { cn } from "lib";
import type { DateRange } from "react-day-picker";
import { format } from "date-fns";
import { useStatuses } from "@/lib/hooks/statuses";
import { StoryStatusIcon } from "./story-status-icon";

export type StoriesFilter = {
  activeSprints: boolean;
  assignedToMe: boolean;
  dueToday: boolean;
  dueThisWeek: boolean;
  completed: boolean;
  startDate: Date | null;
  endDate: Date | null;
  assingee: string[];
  createdFrom: Date | null;
  createdTo: Date | null;
  issueType: string[];
  labels: string[];
  priority: string[];
  createdBy: string[];
  sprint: string[];
  status: string[];
  updatedFrom: Date | null;
  updatedTo: Date | null;
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
  const [date, _] = useState<DateRange | undefined>({
    from: new Date(),
    to: new Date(),
  });

  const { data: statuses = [] } = useStatuses();
  const doneStatusId = statuses.find(
    (state) => state.category === "completed",
  )?.id;

  // filtersCount returns the number of filters applied.
  const filtersCount = () => {
    let count = 0;
    type StoriesFilterKey = keyof StoriesFilter;
    for (const key in filters) {
      const typedKey = key as StoriesFilterKey;
      if (Array.isArray(filters[typedKey])) {
        count += (filters[typedKey] as unknown[]).length > 0 ? 1 : 0;
      } else {
        count += filters[typedKey] ? 1 : 0;
      }
    }
    return count;
  };

  const getButtonLabel = () => {
    if (filtersCount()) {
      return `${filtersCount()} filter${filtersCount() > 1 ? "s" : ""} applied`;
    }
    return "Filters";
  };

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          color="tertiary"
          leftIcon={<FilterIcon className="h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          {getButtonLabel()}
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-[26rem] rounded-[0.6rem] pb-5">
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
            icon={<SprintsIcon className="h-5 w-auto" />}
            isActive={filters.activeSprints}
            label="Active sprints"
            onClick={() => {
              setFilters({ ...filters, activeSprints: !filters.activeSprints });
            }}
          />
          <ToggleButton
            icon={<AvatarIcon className="h-5 w-auto" />}
            isActive={filters.assignedToMe}
            label="Assigned to me"
            onClick={() => {
              setFilters({ ...filters, assignedToMe: !filters.assignedToMe });
            }}
          />
          <ToggleButton
            icon={<CalendarIcon className="h-5 w-auto" />}
            isActive={filters.dueToday}
            label="Due today"
            onClick={() => {
              setFilters({ ...filters, dueToday: !filters.dueToday });
            }}
          />
          <ToggleButton
            icon={<CalendarIcon className="h-5 w-auto" />}
            isActive={filters.dueThisWeek}
            label="Due this week"
            onClick={() => {
              setFilters({ ...filters, dueThisWeek: !filters.dueThisWeek });
            }}
          />
          <ToggleButton
            icon={
              <StoryStatusIcon
                className="h-5 w-auto text-dark dark:text-gray-200"
                statusId={doneStatusId}
              />
            }
            isActive={filters.completed}
            label="Completed"
            onClick={() => {
              setFilters({ ...filters, completed: !filters.completed });
            }}
          />
          <Divider />

          <Box className="mt-4 px-4">
            <Text
              color="muted"
              fontSize="sm"
              fontWeight="semibold"
              transform="uppercase"
            >
              Date Range
            </Text>
            <Flex align="end" className="mt-2">
              <Box className="w-full">
                <Text className="mb-1">Start date</Text>
                <DatePicker>
                  <DatePicker.Trigger>
                    <Button
                      color="tertiary"
                      fullWidth
                      leftIcon={
                        <CalendarIcon className="relative -top-[0.5px] h-[1.1rem] w-auto" />
                      }
                      variant="outline"
                    >
                      {format(new Date(), "LLL dd, y")}
                    </Button>
                  </DatePicker.Trigger>
                  <DatePicker.Calendar
                    defaultMonth={date?.from}
                    initialFocus
                    // onSelect={handleSearch}
                    selected={new Date()}
                  />
                </DatePicker>
              </Box>
              <Box className="flex h-[2.4rem] items-center px-2 md:h-[2.5rem]">
                <ArrowRightIcon className="h-[1.15rem] w-auto" />
              </Box>
              <Box className="w-full">
                <Text className="mb-1">Deadline</Text>
                <DatePicker>
                  <DatePicker.Trigger>
                    <Button
                      color="tertiary"
                      fullWidth
                      leftIcon={
                        <CalendarIcon className="relative -top-[0.5px] h-[1.1rem] w-auto" />
                      }
                      variant="outline"
                    >
                      {format(new Date(), "LLL dd, y")}
                    </Button>
                  </DatePicker.Trigger>
                  <DatePicker.Calendar
                    initialFocus
                    mode="range"
                    // selected={date}

                    numberOfMonths={2}
                  />
                </DatePicker>
              </Box>
            </Flex>

            <Box className="my-6">
              <Text
                className="mb-2"
                color="muted"
                fontSize="sm"
                fontWeight="semibold"
                transform="uppercase"
              >
                Assignee
              </Text>

              <Flex className="gap-x-1.5 gap-y-2" justify="center" wrap>
                {new Array(20).fill(0).map((_, i) => (
                  <Avatar key={i} name="John Doe" />
                ))}
              </Flex>
            </Box>
            <Text
              color="muted"
              fontSize="sm"
              fontWeight="semibold"
              transform="uppercase"
            >
              Created
            </Text>
            <Flex align="end" className="mt-2">
              <Box className="w-full">
                <Text className="mb-1">From</Text>
                <DatePicker>
                  <DatePicker.Trigger>
                    <Button
                      color="tertiary"
                      fullWidth
                      leftIcon={
                        <CalendarIcon className="relative -top-[0.5px] h-[1.1rem] w-auto" />
                      }
                      variant="outline"
                    >
                      {format(new Date(), "LLL dd, y")}
                    </Button>
                  </DatePicker.Trigger>
                  <DatePicker.Calendar
                    defaultMonth={date?.from}
                    initialFocus
                    // onSelect={handleSearch}
                    selected={new Date()}
                  />
                </DatePicker>
              </Box>
              <Box className="flex h-[2.4rem] items-center px-2 md:h-[2.5rem]">
                <ArrowRightIcon className="h-[1.15rem] w-auto" />
              </Box>
              <Box className="w-full">
                <Text className="mb-1">To</Text>
                <DatePicker>
                  <DatePicker.Trigger>
                    <Button
                      color="tertiary"
                      fullWidth
                      leftIcon={
                        <CalendarIcon className="relative -top-[0.5px] h-[1.1rem] w-auto" />
                      }
                      variant="outline"
                    >
                      {format(new Date(), "LLL dd, y")}
                    </Button>
                  </DatePicker.Trigger>
                  <DatePicker.Calendar
                    defaultMonth={date?.from}
                    initialFocus
                    // onSelect={handleSearch}
                    selected={new Date()}
                  />
                </DatePicker>
              </Box>
            </Flex>
          </Box>
        </Box>
      </Popover.Content>
    </Popover>
  );
};
