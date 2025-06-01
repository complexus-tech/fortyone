"use client";

import { Box, Flex, Text } from "ui";
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
} from "date-fns";
import { useParams } from "next/navigation";
import { useState, useCallback, useEffect, useRef } from "react";
import type { Story } from "@/modules/stories/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { BodyContainer } from "../shared/body";

// Interactive gantt bar component
const GanttBar = ({
  story,
  dateRange,
  onDateUpdate,
}: {
  story: Story;
  dateRange: { start: Date; end: Date };
  onDateUpdate: (storyId: string, startDate: string, endDate: string) => void;
}) => {
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<{
    x: number;
    type: "move" | "resize-start" | "resize-end";
  } | null>(null);

  // Calculate positions
  const startDate = story.startDate
    ? new Date(story.startDate)
    : new Date(story.createdAt);
  const endDate = story.endDate
    ? new Date(story.endDate)
    : addDays(startDate, 1);

  const totalDays = differenceInDays(dateRange.end, dateRange.start);
  const storyStartOffset = differenceInDays(startDate, dateRange.start);
  const storyDuration = differenceInDays(endDate, startDate) || 1;

  const leftPosition = Math.max(0, (storyStartOffset / totalDays) * 100);
  const width = Math.min(100 - leftPosition, (storyDuration / totalDays) * 100);

  // Event handlers
  const handleMouseDown = useCallback(
    (e: React.MouseEvent, type: "move" | "resize-start" | "resize-end") => {
      e.preventDefault();
      setIsDragging(true);
      setDragStart({ x: e.clientX, type });
    },
    [],
  );

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !dragStart) return;

      const deltaX = e.clientX - dragStart.x;
      const container = (e.target as HTMLElement).closest(".gantt-timeline");
      if (!container) return;

      const containerWidth = container.clientWidth;
      const deltaPercent = (deltaX / containerWidth) * 100;
      const deltaDays = Math.round((deltaPercent / 100) * totalDays);

      if (deltaDays === 0) return;

      let newStartDate = startDate;
      let newEndDate = endDate;

      if (dragStart.type === "move") {
        newStartDate = addDays(startDate, deltaDays);
        newEndDate = addDays(endDate, deltaDays);
      } else if (dragStart.type === "resize-start") {
        newStartDate = addDays(startDate, deltaDays);
        if (newStartDate >= endDate) newStartDate = addDays(endDate, -1);
      } else {
        // resize-end case
        newEndDate = addDays(endDate, deltaDays);
        if (newEndDate <= startDate) newEndDate = addDays(startDate, 1);
      }

      onDateUpdate(story.id, formatISO(newStartDate), formatISO(newEndDate));
      setDragStart({ ...dragStart, x: e.clientX });
    },
    [
      isDragging,
      dragStart,
      startDate,
      endDate,
      totalDays,
      story.id,
      onDateUpdate,
    ],
  );

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
    setDragStart(null);
  }, []);

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

  // Early returns after hooks
  if (!story.startDate && !story.endDate) return null;
  if (width <= 0) return null;

  return (
    <Box
      className={cn(
        "bg-blue-500 hover:bg-blue-600 group absolute h-6 cursor-pointer rounded transition-colors",
        {
          "ring-blue-300 shadow-lg ring-2": isDragging,
        },
      )}
      onMouseDown={(e) => {
        handleMouseDown(e, "move");
      }}
      style={{
        left: `${leftPosition}%`,
        width: `${width}%`,
        top: "6px",
      }}
    >
      {/* Resize handle - start */}
      <Box
        className="bg-blue-700 absolute left-0 top-0 h-full w-1 cursor-ew-resize opacity-0 group-hover:opacity-100"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-start");
        }}
      />

      {/* Resize handle - end */}
      <Box
        className="bg-blue-700 absolute right-0 top-0 h-full w-1 cursor-ew-resize opacity-0 group-hover:opacity-100"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-end");
        }}
      />

      {/* Story title inside bar */}
      {width > 15 && (
        <Text
          className="absolute inset-0 flex items-center truncate px-2 text-xs text-white"
          fontWeight="medium"
        >
          {story.title}
        </Text>
      )}
    </Box>
  );
};

// Story row component
const GanttRow = ({
  story,
  dateRange,
  onDateUpdate,
  teamCode,
}: {
  story: Story;
  dateRange: { start: Date; end: Date };
  onDateUpdate: (storyId: string, startDate: string, endDate: string) => void;
  teamCode?: string;
}) => {
  // Calculate duration
  const startDate = story.startDate ? new Date(story.startDate) : null;
  const endDate = story.endDate ? new Date(story.endDate) : null;
  const duration =
    startDate && endDate ? differenceInDays(endDate, startDate) : null;

  const days = eachDayOfInterval({
    start: dateRange.start,
    end: dateRange.end,
  });

  // Calculate minimum width for timeline to ensure proper sticky behavior
  const timelineMinWidth = days.length * 64; // 64px per day (min-w-16 = 64px)
  const totalRowWidth = 30 * 16 + 24 * 4 + timelineMinWidth; // 30rem + 6rem + timeline width (converted to px)

  return (
    <Flex
      align="center"
      className="group relative h-14 hover:bg-gray-100/60 dark:hover:bg-dark-200/60"
      style={{ minWidth: `${totalRowWidth}px` }}
    >
      {/* Combined sticky story info and duration column */}
      <Flex
        align="center"
        className="sticky left-0 z-10 h-full w-[34rem] shrink-0 border-r border-gray-200 bg-white px-6 group-hover:bg-gray-100/60 dark:border-dark-200 dark:bg-dark dark:group-hover:bg-dark-200/60"
        justify="between"
      >
        {/* Story info */}
        <Flex align="center" className="min-w-0 flex-1 gap-1.5">
          <Text className="w-16 shrink-0 text-[0.9rem]" color="muted">
            {teamCode}-{story.sequenceId}
          </Text>
          <Text className="line-clamp-1">{story.title}</Text>
        </Flex>

        {/* Duration (only if exists) */}
        {duration ? (
          <Text className="ml-4 shrink-0" color="muted">
            {duration} day{duration !== 1 ? "s" : ""}
          </Text>
        ) : null}
      </Flex>

      {/* Gantt chart area with vertical grid lines */}
      <Box
        className="relative h-full"
        style={{ minWidth: `${timelineMinWidth}px` }}
      >
        {/* Vertical grid lines for each day */}
        <Flex className="absolute inset-0">
          {days.map((day) => (
            <Box
              className={cn(
                "min-w-16 flex-1 border-r border-gray-100 dark:border-dark-200",
                {
                  "bg-gray-50/50 dark:bg-dark-200/20": isWeekend(day),
                },
              )}
              key={day.getTime()}
            />
          ))}
        </Flex>

        {/* Gantt bar on top of grid */}
        <Box className="relative z-10 h-full px-2">
          <GanttBar
            dateRange={dateRange}
            onDateUpdate={onDateUpdate}
            story={story}
          />
        </Box>
      </Box>
    </Flex>
  );
};

// Timeline header component
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
  const totalRowWidth = 30 * 16 + 24 * 4 + timelineMinWidth; // 30rem + 6rem + timeline width (converted to px)

  return (
    <Box
      className="sticky top-0 z-20 border-b border-r border-gray-200 dark:border-dark-200"
      style={{ minWidth: `${totalRowWidth}px` }}
    >
      <Flex>
        {/* Combined sticky stories/duration header */}
        <Flex
          align="center"
          className="sticky left-0 z-10 w-[34rem] shrink-0 border-r border-gray-200 bg-white px-6 py-2.5 dark:border-dark-200 dark:bg-dark"
          justify="between"
        >
          <Text color="muted" fontWeight="medium">
            Stories
          </Text>
          <Text color="muted" fontWeight="medium">
            Duration
          </Text>
        </Flex>

        {/* Timeline header - show week/month spans */}
        <Flex
          className="bg-white dark:bg-dark"
          style={{ minWidth: `${timelineMinWidth}px` }}
        >
          <Box className="w-full">
            {/* Week row */}
            <Box className="border-b border-gray-100 dark:border-dark-200">
              <Flex>
                {getWeekSpans(days).map(({ week, month, span }, index) => (
                  <Box
                    className="border-r border-gray-100 px-2 py-1.5 text-left dark:border-dark-200"
                    key={`${month}-${week}-${index}`}
                    style={{ width: `${(span / days.length) * 100}%` }}
                  >
                    <Flex
                      align="center"
                      className="h-5 min-h-0"
                      justify="between"
                    >
                      <Text color="muted" fontSize="sm" fontWeight="medium">
                        {month}
                      </Text>
                      <Text
                        className="opacity-80"
                        color="muted"
                        fontSize="sm"
                        fontWeight="medium"
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
              {days.map((day) => (
                <Box
                  className={cn(
                    "min-w-16 flex-1 border-r border-gray-100 px-1 py-1 text-center dark:border-dark-200",
                    {
                      "bg-gray-50 dark:bg-dark-200/30": isWeekend(day),
                    },
                  )}
                  key={day.getTime()}
                >
                  <Flex align="center" className="px-1" justify="between">
                    <Text color="muted" fontSize="sm">
                      {format(day, "d")}
                    </Text>
                    <Text color="muted" fontSize="sm">
                      {format(day, "eeeee")}
                    </Text>
                  </Flex>
                </Box>
              ))}
            </Flex>
          </Box>
        </Flex>
      </Flex>
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
  return {
    start: subDays(centerDate, halfViewport),
    end: addDays(centerDate, halfViewport),
  };
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

  const team = teams.find(({ id }) => id === teamId);
  const teamCode = team?.code || "STORY";

  // Calculate visible date range centered on today (1 year total)
  const dateRange = getVisibleDateRange(new Date(), 365);

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

  // Auto-scroll to center on today when component mounts
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // Wait for next frame to ensure everything is rendered
    const timer = requestAnimationFrame(() => {
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

    return () => {
      cancelAnimationFrame(timer);
    };
  }, [dateRange.start, dateRange.end]);

  // Filter stories to only show those with dates
  const storiesWithDates = stories.filter(
    (story) => story.startDate || story.endDate,
  );

  if (storiesWithDates.length === 0) {
    return (
      <BodyContainer
        className={cn("flex h-96 items-center justify-center", className)}
      >
        <Box className="max-w-lg text-center">
          <Text className="mb-4" fontSize="xl" fontWeight="semibold">
            No stories with dates
          </Text>
          <Text color="muted">
            Add start dates or deadlines to stories to see them in the gantt
            chart. Once visible, you can drag the bars to update dates
            interactively.
          </Text>
        </Box>
      </BodyContainer>
    );
  }

  return (
    <div
      className={cn(
        "relative left-px overflow-x-auto overflow-y-auto",
        className,
      )}
      ref={containerRef}
    >
      <TimelineHeader dateRange={dateRange} />
      <Box className="min-w-max">
        {storiesWithDates.map((story) => (
          <GanttRow
            dateRange={dateRange}
            key={story.id}
            onDateUpdate={handleDateUpdate}
            story={story}
            teamCode={teamCode}
          />
        ))}
      </Box>
    </div>
  );
};
