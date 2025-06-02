"use client";

import { Box, Button, Flex, Menu, Text } from "ui";
import { cn } from "lib";
import {
  format,
  eachDayOfInterval,
  addDays,
  differenceInDays,
  formatISO,
  isWeekend,
  subDays,
  getWeek,
  isSameWeek,
  isYesterday,
} from "date-fns";
import { useParams } from "next/navigation";
import { useState, useCallback, useEffect, useRef } from "react";
import { ArrowDown2Icon } from "icons";
import type { Story } from "@/modules/stories/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { StatusesMenu } from "@/components/ui/story/statuses-menu";
import { PriorityIcon } from "./priority-icon";
import { StoryStatusIcon } from "./story-status-icon";

// Interactive gantt bar component
const GanttBar = ({
  story,
  dateRange,
  onDateUpdate,
  containerRef,
}: {
  story: Story;
  dateRange: { start: Date; end: Date };
  onDateUpdate: (storyId: string, startDate: string, endDate: string) => void;
  containerRef: React.RefObject<HTMLDivElement | null>;
}) => {
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<{
    x: number;
    type: "move" | "resize-start" | "resize-end";
    originalStartDate: Date;
    originalEndDate: Date;
  } | null>(null);

  // Local state for visual feedback during drag
  const [dragPosition, setDragPosition] = useState<{
    leftPosition: number;
    width: number;
  } | null>(null);

  // Calculate positions (will be used only if both dates exist)
  const startDate = story.startDate ? new Date(story.startDate) : new Date();
  const endDate = story.endDate
    ? new Date(story.endDate)
    : addDays(startDate, 1);

  // Normalize dates to start of day to ensure proper alignment
  startDate.setHours(0, 0, 0, 0);
  endDate.setHours(0, 0, 0, 0);

  const storyStartOffset = differenceInDays(startDate, dateRange.start);
  const storyDuration = differenceInDays(endDate, startDate) || 1;

  // Use pixel-based positioning to align with 64px date columns
  const dayWidth = 64; // Each day column is 64px wide (min-w-16)
  const leftPosition =
    dragPosition?.leftPosition ?? storyStartOffset * dayWidth;
  const width = dragPosition?.width ?? storyDuration * dayWidth;

  // Event handlers
  const handleMouseDown = useCallback(
    (e: React.MouseEvent, type: "move" | "resize-start" | "resize-end") => {
      e.preventDefault();
      setIsDragging(true);
      setDragStart({
        x: e.clientX,
        type,
        originalStartDate: startDate,
        originalEndDate: endDate,
      });
    },
    [startDate, endDate],
  );

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !dragStart) return;

      const deltaX = e.clientX - dragStart.x;
      const container = containerRef.current;
      if (!container) return;

      // Calculate how many days we've moved based on pixel movement
      const deltaDays = Math.round(deltaX / dayWidth);

      if (deltaDays === 0) return;

      let newStartDate = dragStart.originalStartDate;
      let newEndDate = dragStart.originalEndDate;

      if (dragStart.type === "move") {
        newStartDate = addDays(dragStart.originalStartDate, deltaDays);
        newEndDate = addDays(dragStart.originalEndDate, deltaDays);
      } else if (dragStart.type === "resize-start") {
        newStartDate = addDays(dragStart.originalStartDate, deltaDays);
        if (newStartDate >= dragStart.originalEndDate) {
          newStartDate = addDays(dragStart.originalEndDate, -1);
        }
      } else {
        // resize-end case
        newEndDate = addDays(dragStart.originalEndDate, deltaDays);
        if (newEndDate <= dragStart.originalStartDate) {
          newEndDate = addDays(dragStart.originalStartDate, 1);
        }
      }

      // Update visual position only
      const newStartOffset = differenceInDays(newStartDate, dateRange.start);
      const newDuration = differenceInDays(newEndDate, newStartDate) || 1;
      const newLeftPosition = newStartOffset * dayWidth;
      const newWidth = newDuration * dayWidth;

      setDragPosition({
        leftPosition: newLeftPosition,
        width: newWidth,
      });
    },
    [isDragging, dragStart, dayWidth, dateRange.start],
  );

  const handleMouseUp = useCallback(() => {
    if (!isDragging || !dragStart || !dragPosition) {
      setIsDragging(false);
      setDragStart(null);
      setDragPosition(null);
      return;
    }

    // Calculate final dates from drag position
    const finalStartOffset = Math.round(dragPosition.leftPosition / dayWidth);
    const finalDuration = Math.round(dragPosition.width / dayWidth);

    const finalStartDate = addDays(dateRange.start, finalStartOffset);
    const finalEndDate = addDays(finalStartDate, Math.max(1, finalDuration));

    // Update story dates only on drag end
    onDateUpdate(story.id, formatISO(finalStartDate), formatISO(finalEndDate));

    // Reset drag state
    setIsDragging(false);
    setDragStart(null);
    setDragPosition(null);
  }, [
    isDragging,
    dragStart,
    dragPosition,
    dayWidth,
    dateRange.start,
    story.id,
    onDateUpdate,
  ]);

  // Global mouse event listeners
  useEffect(() => {
    if (isDragging) {
      document.addEventListener("mousemove", handleMouseMove);
      document.addEventListener("mouseup", handleMouseUp);
      return () => {
        document.removeEventListener("mousemove", handleMouseMove);
        document.removeEventListener("mouseup", handleMouseUp);
      };
    }
  }, [isDragging, handleMouseMove, handleMouseUp]);

  // Early return if story doesn't have both start and end dates (after all hooks)
  if (!story.startDate || !story.endDate) return null;
  if (width <= 0) return null;

  return (
    <Box
      className={cn(
        "group absolute z-0 h-10 cursor-pointer overflow-hidden rounded-lg border-[0.5px] border-gray-200/60 bg-gray-100 transition-colors dark:border-dark-50/80 dark:bg-dark-200",
        {
          "shadow-lg ring-2": isDragging,
        },
      )}
      onMouseDown={(e) => {
        handleMouseDown(e, "move");
      }}
      style={{
        left: `${leftPosition}px`,
        width: `${width}px`,
        top: "6px",
      }}
    >
      {/* Resize handle - start */}
      <Box
        className="absolute left-0 top-0 h-full w-1 cursor-ew-resize"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-start");
        }}
      />

      {/* Resize handle - end */}
      <Box
        className="absolute right-0 top-0 h-full w-1 cursor-ew-resize"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-end");
        }}
      />

      {/* Story title inside bar */}

      <Text
        className="line-clamp-1 block truncate px-3 leading-10"
        fontWeight="medium"
      >
        {story.title}
      </Text>
    </Box>
  );
};

// Timeline header component - now only handles the date columns
const TimelineHeader = ({
  dateRange,
}: {
  dateRange: { start: Date; end: Date };
}) => {
  const days = eachDayOfInterval({
    start: dateRange.start,
    end: dateRange.end,
  });

  // Calculate minimum width for timeline to ensure proper sticky behavior
  const timelineMinWidth = days.length * 64; // 64px per day (min-w-16 = 64px)

  // Get today's date for comparison
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  return (
    <Box
      className="sticky top-0 z-10 h-16 border-b-[0.5px] border-gray-200/60 bg-white dark:border-dark-100 dark:bg-dark"
      style={{ minWidth: `${timelineMinWidth}px` }}
    >
      <Box className="h-8 w-full">
        {/* Week row */}
        <Box className="border-b-[0.5px] border-gray-100 dark:border-dark-100">
          <Flex>
            {getWeekSpans(days).map(({ week, month, span }, index) => (
              <Box
                className="border-r-[0.5px] border-gray-100 px-2 py-1.5 text-left dark:border-dark-100"
                key={`${month}-${week}-${index}`}
                style={{ width: `${(span / days.length) * 100}%` }}
              >
                <Flex align="center" className="h-5 min-h-0" justify="between">
                  <Text
                    className="text-[0.9rem]"
                    color="muted"
                    fontWeight="semibold"
                  >
                    {month}
                  </Text>
                  <Text
                    className="text-[0.9rem] opacity-60"
                    color="muted"
                    fontWeight="semibold"
                  >
                    {week}
                  </Text>
                </Flex>
              </Box>
            ))}
          </Flex>
        </Box>

        {/* Days row - compact */}
        <Flex>
          {days.map((day) => {
            const isToday = day.getTime() === today.getTime();

            return (
              <Box
                className={cn(
                  "h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] border-gray-100 px-1 py-1 text-center dark:border-dark-100",
                  {
                    "bg-gray-50 dark:bg-dark-200/30":
                      isWeekend(day) && !isToday,
                    "border-primary bg-primary dark:border-primary": isToday,
                  },
                )}
                key={day.getTime()}
              >
                <Flex align="center" className="px-1" justify="between">
                  {isToday ? (
                    <Text color="white" fontSize="sm" fontWeight="medium">
                      Today
                    </Text>
                  ) : (
                    <>
                      <Text color="muted" fontSize="sm">
                        {format(day, "d")}
                      </Text>
                      <Text color="muted" fontSize="sm">
                        {format(day, "eeeee")}
                      </Text>
                    </>
                  )}
                </Flex>
              </Box>
            );
          })}
        </Flex>
      </Box>
    </Box>
  );
};

// Stories header component
const StoriesHeader = ({ onReset }: { onReset: () => void }) => {
  return (
    <Flex
      align="center"
      className="sticky top-0 h-16 border-b-[0.5px] border-gray-200/60 bg-white px-6 py-2.5 dark:border-dark-100 dark:bg-dark"
      justify="between"
    >
      <Button color="tertiary" onClick={onReset} size="sm">
        Today
      </Button>
      <Flex align="center" gap={2}>
        <Text color="muted" fontWeight="medium">
          Zoom:
        </Text>
        <Menu>
          <Menu.Button>
            <Button
              color="tertiary"
              onClick={onReset}
              rightIcon={<ArrowDown2Icon className="h-4" />}
              size="sm"
            >
              Weeks
            </Button>
          </Menu.Button>
          <Menu.Items className="w-40">
            <Menu.Group>
              <Menu.Item>Weeks</Menu.Item>
              <Menu.Item>Months</Menu.Item>
              <Menu.Item>Quarters</Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Flex>
  );
};

// Sticky stories section component
const StoriesSection = ({
  stories,
  teamCode,
  onReset,
}: {
  stories: Story[];
  teamCode: string;
  onReset: () => void;
}) => {
  return (
    <Box className="sticky left-0 z-20 w-[34rem] shrink-0 border-r-[0.5px] border-gray-200/60 bg-white dark:border-dark-100 dark:bg-dark">
      <StoriesHeader onReset={onReset} />
      {stories.map((story) => {
        // Calculate duration
        const startDate = story.startDate ? new Date(story.startDate) : null;
        const endDate = story.endDate ? new Date(story.endDate) : null;
        const duration =
          startDate && endDate ? differenceInDays(endDate, startDate) : null;

        return (
          <Flex
            align="center"
            className="group h-14 border-b-[0.5px] border-gray-100 px-6 transition-colors hover:bg-gray-50 dark:border-dark-100 dark:hover:bg-dark-300"
            justify="between"
            key={story.id}
          >
            {/* Story info */}
            <Flex align="center" className="min-w-0 flex-1 gap-2">
              <Text
                className="line-clamp-1 w-16 shrink-0 text-[0.95rem]"
                color="muted"
              >
                {teamCode}-{story.sequenceId}
              </Text>
              <PrioritiesMenu>
                <PrioritiesMenu.Trigger>
                  <button
                    className="flex shrink-0 select-none items-center gap-1 disabled:cursor-not-allowed disabled:opacity-50"
                    type="button"
                  >
                    <PriorityIcon priority={story.priority} />
                    <span className="sr-only">{story.priority}</span>
                  </button>
                </PrioritiesMenu.Trigger>
                <PrioritiesMenu.Items
                  priority={story.priority}
                  setPriority={(_p) => {
                    // handleUpdate({ priority: p });
                  }}
                />
              </PrioritiesMenu>
              <StatusesMenu>
                <StatusesMenu.Trigger>
                  <button
                    className="flex items-center gap-1 disabled:cursor-not-allowed disabled:opacity-50"
                    type="button"
                  >
                    <StoryStatusIcon statusId={story.statusId} />
                    <span className="sr-only">Story status</span>
                  </button>
                </StatusesMenu.Trigger>
                <StatusesMenu.Items
                  setStatusId={(_statusId) => {
                    // handleUpdate({ statusId });
                  }}
                  statusId={story.statusId}
                  teamId={story.teamId}
                />
              </StatusesMenu>
              <Text className="line-clamp-1">{story.title}</Text>
            </Flex>

            {/* Duration (only if exists) */}
            {duration ? (
              <Text className="ml-4 shrink-0" color="muted">
                {duration} day{duration !== 1 ? "s" : ""}
              </Text>
            ) : null}
          </Flex>
        );
      })}
    </Box>
  );
};

// Chart section component
const ChartSection = ({
  stories,
  dateRange,
  onDateUpdate,
  containerRef,
}: {
  stories: Story[];
  dateRange: { start: Date; end: Date };
  onDateUpdate: (storyId: string, startDate: string, endDate: string) => void;
  containerRef: React.RefObject<HTMLDivElement | null>;
}) => {
  const days = eachDayOfInterval({
    start: dateRange.start,
    end: dateRange.end,
  });

  // Calculate minimum width for timeline
  const timelineMinWidth = days.length * 64; // 64px per day

  // Get today's date for comparison
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  return (
    <Box className="flex-1" style={{ minWidth: `${timelineMinWidth}px` }}>
      <TimelineHeader dateRange={dateRange} />
      {stories.map((story) => (
        <Box className="relative h-14" key={story.id}>
          {/* Vertical grid lines for each day */}
          <Flex className="absolute inset-0">
            {days.map((day) => {
              const isToday = day.getTime() === today.getTime();
              const dayIsYesterday = isYesterday(day);

              return (
                <Box
                  className={cn(
                    "min-w-16 flex-1 border-r-[0.5px] border-gray-100 dark:border-dark-100",
                    {
                      "bg-gray-50/50 dark:bg-dark-200/20":
                        isWeekend(day) && !isToday,
                      "border-primary/50 bg-primary/10 dark:border-primary/50":
                        isToday,
                      "border-primary/50 dark:border-primary/50":
                        dayIsYesterday,
                    },
                  )}
                  key={day.getTime()}
                />
              );
            })}
          </Flex>

          {/* Gantt bar on top of grid */}
          <Box className="z-5 relative h-full px-2">
            <GanttBar
              containerRef={containerRef}
              dateRange={dateRange}
              onDateUpdate={onDateUpdate}
              story={story}
            />
          </Box>
        </Box>
      ))}
    </Box>
  );
};

// Helper function to get week spans
const getWeekSpans = (days: Date[]) => {
  if (days.length === 0) return [];

  const spans: {
    week: string;
    month: string;
    startIndex: number;
    span: number;
  }[] = [];
  let startIndex = 0;

  for (let i = 0; i < days.length; i++) {
    const currentDay = days[i];
    const nextDay = days[i + 1];

    // Check if this is the last day of the current week or the last day overall
    const isEndOfWeek =
      i === days.length - 1 ||
      !isSameWeek(currentDay, nextDay, { weekStartsOn: 0 }); // 0 = Sunday

    if (isEndOfWeek) {
      const span = i - startIndex + 1;
      const weekStart = days[startIndex];
      const weekNumber = getWeek(weekStart, { weekStartsOn: 0 });
      const monthYear = format(weekStart, "MMM yyyy");

      spans.push({
        week: `Week ${weekNumber}`,
        month: monthYear,
        startIndex,
        span,
      });

      startIndex = i + 1;
    }
  }

  return spans;
};

// Helper to get the visible date range for the gantt chart
const getVisibleDateRange = (centerDate: Date, viewportDays = 365) => {
  const halfViewport = Math.floor(viewportDays / 2);
  const start = subDays(centerDate, halfViewport);
  const end = addDays(centerDate, halfViewport);

  // Normalize to start of day for consistent alignment
  start.setHours(0, 0, 0, 0);
  end.setHours(0, 0, 0, 0);

  return { start, end };
};

export const GanttBoard = ({
  stories,
  className,
}: {
  stories: Story[];
  className?: string;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const { mutate } = useUpdateStoryMutation();
  const containerRef = useRef<HTMLDivElement>(null);
  const hasScrolledRef = useRef(false);

  const team = teams.find(({ id }) => id === teamId);
  const teamCode = team?.code || "STORY";

  // Calculate visible date range centered on today (1 year total)
  const today = new Date();
  today.setHours(0, 0, 0, 0); // Normalize to start of day
  const dateRange = getVisibleDateRange(today, 365);

  // Handle date updates from drag operations
  const handleDateUpdate = useCallback(
    (storyId: string, startDate: string, endDate: string) => {
      mutate({
        storyId,
        payload: {
          startDate,
          endDate,
        },
      });
    },
    [mutate],
  );

  // Scroll to today function - reusable for both initial scroll and reset
  const scrollToToday = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    // Wait for next frame to ensure everything is rendered
    requestAnimationFrame(() => {
      const now = new Date();
      const daysFromStart = differenceInDays(now, dateRange.start);

      // Calculate position: sticky columns width + (current date position * day width)
      const stickyColumnsWidth = 34 * 16; // 34rem converted to px
      const dayWidth = 64; // min-w-16 = 64px
      const currentDatePixelPosition = daysFromStart * dayWidth;

      // Center the current date in the viewport
      const viewportWidth = container.clientWidth;
      const scrollPosition =
        stickyColumnsWidth + currentDatePixelPosition - viewportWidth / 2;

      container.scrollLeft = Math.max(0, scrollPosition);
    });
  }, [dateRange.start]);

  // Auto-scroll to center on today when component mounts (only once)
  useEffect(() => {
    if (!hasScrolledRef.current) {
      scrollToToday();
      hasScrolledRef.current = true;
    }
  }, [scrollToToday]);

  // Filter stories to only show those with dates
  const storiesWithDates = stories.filter(
    (story) => story.startDate || story.endDate,
  );

  return (
    <div
      className={cn(
        "relative left-px overflow-x-auto overflow-y-auto",
        className,
      )}
      ref={containerRef}
    >
      <Flex className="min-w-max">
        <StoriesSection
          onReset={scrollToToday}
          stories={storiesWithDates}
          teamCode={teamCode}
        />
        <ChartSection
          containerRef={containerRef}
          dateRange={dateRange}
          onDateUpdate={handleDateUpdate}
          stories={storiesWithDates}
        />
      </Flex>
    </div>
  );
};
