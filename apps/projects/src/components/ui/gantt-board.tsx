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
  eachMonthOfInterval,
  eachQuarterOfInterval,
  startOfMonth,
  endOfMonth,
  startOfQuarter,
  endOfQuarter,
  addMonths,
  addQuarters,
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

// Types
type ZoomLevel = "weeks" | "months" | "quarters";

// Helper functions
const getTimePeriodsForZoom = (
  dateRange: { start: Date; end: Date },
  zoomLevel: ZoomLevel,
) => {
  switch (zoomLevel) {
    case "weeks":
      return eachDayOfInterval({
        start: dateRange.start,
        end: dateRange.end,
      });
    case "months":
      return eachMonthOfInterval({
        start: startOfMonth(dateRange.start),
        end: endOfMonth(dateRange.end),
      });
    case "quarters":
      return eachQuarterOfInterval({
        start: startOfQuarter(dateRange.start),
        end: endOfQuarter(dateRange.end),
      });
    default:
      return [];
  }
};

const getColumnWidth = (zoomLevel: ZoomLevel) => {
  switch (zoomLevel) {
    case "weeks":
      return 64; // 64px per day
    case "months":
      return 120; // 120px per month
    case "quarters":
      return 180; // 180px per quarter
    default:
      return 64;
  }
};

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

const getVisibleDateRange = (centerDate: Date, viewportDays = 365) => {
  const halfViewport = Math.floor(viewportDays / 2);
  const start = subDays(centerDate, halfViewport);
  const end = addDays(centerDate, halfViewport);

  // Normalize to start of day for consistent alignment
  start.setHours(0, 0, 0, 0);
  end.setHours(0, 0, 0, 0);

  return { start, end };
};

// Bar Component - Interactive gantt bar
const Bar = ({
  story,
  dateRange,
  onDateUpdate,
  containerRef,
  zoomLevel,
  onDragStateChange,
}: {
  story: Story;
  dateRange: { start: Date; end: Date };
  onDateUpdate: (storyId: string, startDate: string, endDate: string) => void;
  containerRef: React.RefObject<HTMLDivElement | null>;
  zoomLevel: ZoomLevel;
  onDragStateChange: (isDragging: boolean) => void;
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

  // Use pixel-based positioning to align with date columns
  const columnWidth = getColumnWidth(zoomLevel);

  // Event handlers
  const handleMouseDown = useCallback(
    (e: React.MouseEvent, type: "move" | "resize-start" | "resize-end") => {
      e.preventDefault();
      e.stopPropagation();
      // Immediately prevent scrolling before drag state is set
      onDragStateChange(true);
      setIsDragging(true);
      setDragStart({
        x: e.clientX,
        type,
        originalStartDate: startDate,
        originalEndDate: endDate,
      });
    },
    [startDate, endDate, onDragStateChange],
  );

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !dragStart) return;

      const deltaX = e.clientX - dragStart.x;
      const container = containerRef.current;
      if (!container) return;

      // Calculate how many units we've moved based on pixel movement and zoom level
      const deltaUnits = Math.round(deltaX / columnWidth);

      if (deltaUnits === 0) return;

      let newStartDate = dragStart.originalStartDate;
      let newEndDate = dragStart.originalEndDate;

      if (dragStart.type === "move") {
        switch (zoomLevel) {
          case "weeks":
            newStartDate = addDays(dragStart.originalStartDate, deltaUnits);
            newEndDate = addDays(dragStart.originalEndDate, deltaUnits);
            break;
          case "months":
            newStartDate = addMonths(dragStart.originalStartDate, deltaUnits);
            newEndDate = addMonths(dragStart.originalEndDate, deltaUnits);
            break;
          case "quarters":
            newStartDate = addQuarters(dragStart.originalStartDate, deltaUnits);
            newEndDate = addQuarters(dragStart.originalEndDate, deltaUnits);
            break;
        }
      } else if (dragStart.type === "resize-start") {
        switch (zoomLevel) {
          case "weeks":
            newStartDate = addDays(dragStart.originalStartDate, deltaUnits);
            if (newStartDate >= dragStart.originalEndDate) {
              newStartDate = addDays(dragStart.originalEndDate, -1);
            }
            break;
          case "months":
            newStartDate = addMonths(dragStart.originalStartDate, deltaUnits);
            if (newStartDate >= dragStart.originalEndDate) {
              newStartDate = addMonths(dragStart.originalEndDate, -1);
            }
            break;
          case "quarters":
            newStartDate = addQuarters(dragStart.originalStartDate, deltaUnits);
            if (newStartDate >= dragStart.originalEndDate) {
              newStartDate = addQuarters(dragStart.originalEndDate, -1);
            }
            break;
        }
      } else {
        // resize-end case
        switch (zoomLevel) {
          case "weeks":
            newEndDate = addDays(dragStart.originalEndDate, deltaUnits);
            if (newEndDate <= dragStart.originalStartDate) {
              newEndDate = addDays(dragStart.originalStartDate, 1);
            }
            break;
          case "months":
            newEndDate = addMonths(dragStart.originalEndDate, deltaUnits);
            if (newEndDate <= dragStart.originalStartDate) {
              newEndDate = addMonths(dragStart.originalStartDate, 1);
            }
            break;
          case "quarters":
            newEndDate = addQuarters(dragStart.originalEndDate, deltaUnits);
            if (newEndDate <= dragStart.originalStartDate) {
              newEndDate = addQuarters(dragStart.originalStartDate, 1);
            }
            break;
        }
      }

      // Calculate visual position based on new dates
      let newLeftPosition = 0;
      let newWidth = columnWidth;

      switch (zoomLevel) {
        case "weeks": {
          const newStartOffset = differenceInDays(
            newStartDate,
            dateRange.start,
          );
          const newDuration = differenceInDays(newEndDate, newStartDate) || 1;
          newLeftPosition = newStartOffset * columnWidth;
          newWidth = newDuration * columnWidth;
          break;
        }
        case "months":
        case "quarters": {
          // For months/quarters, recalculate position using the same logic as the main positioning
          const periods = getTimePeriodsForZoom(dateRange, zoomLevel);

          if (zoomLevel === "months") {
            const newStartMonth = startOfMonth(newStartDate);
            const newEndMonth = startOfMonth(newEndDate);

            const startOffset = periods.findIndex(
              (period) => period.getTime() === newStartMonth.getTime(),
            );
            const endOffset = periods.findIndex(
              (period) => period.getTime() === newEndMonth.getTime(),
            );

            if (startOffset !== -1) {
              newLeftPosition = startOffset * columnWidth;
              const duration =
                endOffset !== -1 ? Math.max(1, endOffset - startOffset + 1) : 1;
              newWidth = duration * columnWidth;
            }
          } else {
            const newStartQuarter = startOfQuarter(newStartDate);
            const newEndQuarter = startOfQuarter(newEndDate);

            const startOffset = periods.findIndex(
              (period) => period.getTime() === newStartQuarter.getTime(),
            );
            const endOffset = periods.findIndex(
              (period) => period.getTime() === newEndQuarter.getTime(),
            );

            if (startOffset !== -1) {
              newLeftPosition = startOffset * columnWidth;
              const duration =
                endOffset !== -1 ? Math.max(1, endOffset - startOffset + 1) : 1;
              newWidth = duration * columnWidth;
            }
          }
          break;
        }
      }

      setDragPosition({
        leftPosition: newLeftPosition,
        width: newWidth,
      });
    },
    [isDragging, dragStart, columnWidth, dateRange.start, zoomLevel],
  );

  const handleMouseUp = useCallback(() => {
    if (!isDragging || !dragStart || !dragPosition) {
      setIsDragging(false);
      onDragStateChange(false);
      setDragStart(null);
      setDragPosition(null);
      return;
    }

    // Calculate final dates from drag position based on zoom level
    let finalStartDate = dragStart.originalStartDate;
    let finalEndDate = dragStart.originalEndDate;

    switch (zoomLevel) {
      case "weeks": {
        const finalStartOffset = Math.round(
          dragPosition.leftPosition / columnWidth,
        );
        const finalDuration = Math.round(dragPosition.width / columnWidth);

        finalStartDate = addDays(dateRange.start, finalStartOffset);
        finalEndDate = addDays(finalStartDate, Math.max(1, finalDuration));
        break;
      }
      case "months":
      case "quarters": {
        // For months/quarters, calculate based on the visual position
        const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
        const startPeriodIndex = Math.round(
          dragPosition.leftPosition / columnWidth,
        );
        const durationPeriods = Math.round(dragPosition.width / columnWidth);

        if (startPeriodIndex >= 0 && startPeriodIndex < periods.length) {
          const startPeriod = periods[startPeriodIndex];

          if (zoomLevel === "months") {
            finalStartDate = startOfMonth(startPeriod);
            // If spanning only one month, end within that month
            if (durationPeriods === 1) {
              finalEndDate = endOfMonth(startPeriod);
            } else {
              // If spanning multiple months, add the duration minus 1 and end within the last month
              const lastMonth = addMonths(startPeriod, durationPeriods - 1);
              finalEndDate = endOfMonth(lastMonth);
            }
          } else {
            finalStartDate = startOfQuarter(startPeriod);
            // If spanning only one quarter, end within that quarter
            if (durationPeriods === 1) {
              finalEndDate = endOfQuarter(startPeriod);
            } else {
              // If spanning multiple quarters, add the duration minus 1 and end within the last quarter
              const lastQuarter = addQuarters(startPeriod, durationPeriods - 1);
              finalEndDate = endOfQuarter(lastQuarter);
            }
          }
        }
        break;
      }
    }

    // Only update if dates actually changed (prevents unnecessary API calls on clicks)
    const originalStartISO = formatISO(dragStart.originalStartDate);
    const originalEndISO = formatISO(dragStart.originalEndDate);
    const finalStartISO = formatISO(finalStartDate);
    const finalEndISO = formatISO(finalEndDate);

    if (originalStartISO !== finalStartISO || originalEndISO !== finalEndISO) {
      onDateUpdate(story.id, finalStartISO, finalEndISO);
    }

    // Reset drag state
    setIsDragging(false);
    onDragStateChange(false);
    setDragStart(null);
    setDragPosition(null);
  }, [
    isDragging,
    dragStart,
    dragPosition,
    columnWidth,
    dateRange.start,
    story.id,
    onDateUpdate,
    zoomLevel,
    onDragStateChange,
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

  // Calculate position based on zoom level
  const calculatePosition = () => {
    const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
    const columnWidth = getColumnWidth(zoomLevel);

    let startOffset = 0;
    let duration = 1;

    switch (zoomLevel) {
      case "weeks": {
        startOffset = differenceInDays(startDate, dateRange.start);
        duration = differenceInDays(endDate, startDate) || 1;
        break;
      }
      case "months": {
        // Find which month the start date falls in
        const startMonth = startOfMonth(startDate);
        const endMonth = startOfMonth(endDate);

        startOffset = periods.findIndex(
          (period) => period.getTime() === startMonth.getTime(),
        );

        const endOffset = periods.findIndex(
          (period) => period.getTime() === endMonth.getTime(),
        );

        if (startOffset === -1) startOffset = 0;
        if (endOffset === -1) {
          duration = 1;
        } else {
          duration = Math.max(1, endOffset - startOffset + 1);
        }
        break;
      }
      case "quarters": {
        // Find which quarter the start date falls in
        const startQuarter = startOfQuarter(startDate);
        const endQuarter = startOfQuarter(endDate);

        startOffset = periods.findIndex(
          (period) => period.getTime() === startQuarter.getTime(),
        );

        const endOffset = periods.findIndex(
          (period) => period.getTime() === endQuarter.getTime(),
        );

        if (startOffset === -1) startOffset = 0;
        if (endOffset === -1) {
          duration = 1;
        } else {
          duration = Math.max(1, endOffset - startOffset + 1);
        }
        break;
      }
    }

    return {
      leftPosition: startOffset * columnWidth,
      width: duration * columnWidth,
    };
  };

  const { leftPosition: calculatedLeft, width: calculatedWidth } =
    calculatePosition();
  const finalLeftPosition = dragPosition?.leftPosition ?? calculatedLeft;
  const finalWidth = dragPosition?.width ?? calculatedWidth;

  // Early returns after all hooks - but only for missing data
  if (!story.startDate || !story.endDate) return null;
  if (finalWidth <= 0) return null;

  return (
    <Box
      className={cn(
        "group absolute z-0 h-10 cursor-pointer overflow-hidden rounded-lg border-[0.5px] border-gray-200/60 bg-gray-100 transition-colors dark:border-dark-50/80 dark:bg-dark-200",
        {
          "shadow-lg ring-2": isDragging,
        },
      )}
      onFocus={(e) => {
        e.preventDefault();
        e.currentTarget.blur();
      }}
      onKeyDown={(e) => {
        e.preventDefault();
      }}
      onMouseDown={(e) => {
        handleMouseDown(e, "move");
      }}
      style={{
        left: `${finalLeftPosition}px`,
        width: `${finalWidth}px`,
        top: "6px",
      }}
      tabIndex={-1}
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

// Timeline Header Component
const TimelineHeader = ({
  dateRange,
  zoomLevel,
}: {
  dateRange: { start: Date; end: Date };
  zoomLevel: ZoomLevel;
}) => {
  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const columnWidth = getColumnWidth(zoomLevel);

  // Calculate minimum width for timeline to ensure proper sticky behavior
  const timelineMinWidth = periods.length * columnWidth;

  // Get today's date for comparison
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const renderPeriodHeader = () => {
    switch (zoomLevel) {
      case "weeks":
        // Existing week view logic
        return (
          <>
            {/* Week row */}
            <Box className="border-b-[0.5px] border-gray-100 dark:border-dark-100">
              <Flex>
                {getWeekSpans(periods).map(({ week, month, span }, index) => (
                  <Box
                    className="border-r-[0.5px] border-gray-100 px-2 py-1.5 text-left dark:border-dark-100"
                    key={`${month}-${week}-${index}`}
                    style={{ width: `${(span / periods.length) * 100}%` }}
                  >
                    <Flex
                      align="center"
                      className="h-5 min-h-0"
                      justify="between"
                    >
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
              {periods.map((day) => {
                const isToday = day.getTime() === today.getTime();

                return (
                  <Box
                    className={cn(
                      "h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] border-gray-100 px-1 py-1 text-center dark:border-dark-100",
                      {
                        "bg-gray-50 dark:bg-dark-200/30":
                          isWeekend(day) && !isToday,
                        "border-primary bg-primary dark:border-primary":
                          isToday,
                      },
                    )}
                    key={day.getTime()}
                    style={{ minWidth: `${columnWidth}px` }}
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
          </>
        );
      case "months":
        // Monthly view
        return (
          <>
            {/* Month row */}
            <Box className="border-b-[0.5px] border-gray-100 dark:border-dark-100">
              <Flex>
                {periods.map((month) => (
                  <Box
                    className="border-r-[0.5px] border-gray-100 px-2 py-1.5 text-left dark:border-dark-100"
                    key={month.getTime()}
                    style={{ minWidth: `${columnWidth}px` }}
                  >
                    <Flex
                      align="center"
                      className="h-5 min-h-0"
                      justify="between"
                    >
                      <Text
                        className="text-[0.9rem]"
                        color="muted"
                        fontWeight="semibold"
                      >
                        {format(month, "MMM")}
                      </Text>
                      <Text
                        className="text-[0.9rem] opacity-60"
                        color="muted"
                        fontWeight="semibold"
                      >
                        {format(month, "yyyy")}
                      </Text>
                    </Flex>
                  </Box>
                ))}
              </Flex>
            </Box>

            {/* Days row - first and last day */}
            <Flex>
              {periods.map((month) => (
                <Box
                  className="h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] border-gray-100 px-1 py-1 text-center dark:border-dark-100"
                  key={month.getTime()}
                  style={{ minWidth: `${columnWidth}px` }}
                >
                  <Flex align="center" className="px-1" justify="between">
                    <Text color="muted" fontSize="sm">
                      {format(month, "d")}
                    </Text>
                    <Text color="muted" fontSize="sm">
                      {format(endOfMonth(month), "d")}
                    </Text>
                  </Flex>
                </Box>
              ))}
            </Flex>
          </>
        );
      case "quarters":
        // Quarterly view
        return (
          <>
            {/* Quarter row */}
            <Box className="border-b-[0.5px] border-gray-100 dark:border-dark-100">
              <Flex>
                {periods.map((quarter) => (
                  <Box
                    className="border-r-[0.5px] border-gray-100 px-2 py-1.5 text-left dark:border-dark-100"
                    key={quarter.getTime()}
                    style={{ minWidth: `${columnWidth}px` }}
                  >
                    <Flex
                      align="center"
                      className="h-5 min-h-0"
                      justify="between"
                    >
                      <Text
                        className="text-[0.9rem]"
                        color="muted"
                        fontWeight="semibold"
                      >
                        Q{Math.ceil((quarter.getMonth() + 1) / 3)}
                      </Text>
                      <Text
                        className="text-[0.9rem] opacity-60"
                        color="muted"
                        fontWeight="semibold"
                      >
                        {format(quarter, "yyyy")}
                      </Text>
                    </Flex>
                  </Box>
                ))}
              </Flex>
            </Box>

            {/* Months row - first and last month of quarter */}
            <Flex>
              {periods.map((quarter) => {
                const quarterStart = startOfQuarter(quarter);
                const quarterEnd = endOfQuarter(quarter);

                return (
                  <Box
                    className="h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] border-gray-100 px-1 py-1 text-center dark:border-dark-100"
                    key={quarter.getTime()}
                    style={{ minWidth: `${columnWidth}px` }}
                  >
                    <Flex align="center" className="px-1" justify="between">
                      <Text color="muted" fontSize="sm">
                        {format(quarterStart, "MMM")}
                      </Text>
                      <Text color="muted" fontSize="sm">
                        {format(quarterEnd, "MMM")}
                      </Text>
                    </Flex>
                  </Box>
                );
              })}
            </Flex>
          </>
        );
      default:
        return null;
    }
  };

  return (
    <Box
      className="sticky top-0 z-10 h-16 border-b-[0.5px] border-gray-200/60 bg-white dark:border-dark-100 dark:bg-dark"
      style={{ minWidth: `${timelineMinWidth}px` }}
    >
      <Box className="h-8 w-full">{renderPeriodHeader()}</Box>
    </Box>
  );
};

// Header Component
const Header = ({
  onReset,
  zoomLevel,
  onZoomChange,
}: {
  onReset: () => void;
  zoomLevel: ZoomLevel;
  onZoomChange: (zoom: ZoomLevel) => void;
}) => {
  const getZoomLabel = (zoom: ZoomLevel) => {
    switch (zoom) {
      case "weeks":
        return "Weeks";
      case "months":
        return "Months";
      case "quarters":
        return "Quarters";
      default:
        return "Weeks";
    }
  };

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
              rightIcon={<ArrowDown2Icon className="h-4" />}
              size="sm"
            >
              {getZoomLabel(zoomLevel)}
            </Button>
          </Menu.Button>
          <Menu.Items className="w-40">
            <Menu.Group>
              <Menu.Item
                onClick={() => {
                  onZoomChange("weeks");
                }}
              >
                Weeks
              </Menu.Item>
              <Menu.Item
                onClick={() => {
                  onZoomChange("months");
                }}
              >
                Months
              </Menu.Item>
              <Menu.Item
                onClick={() => {
                  onZoomChange("quarters");
                }}
              >
                Quarters
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Flex>
  );
};

// Stories Component
const Stories = ({
  stories,
  teamCode,
  onReset,
  zoomLevel,
  onZoomChange,
}: {
  stories: Story[];
  teamCode: string;
  onReset: () => void;
  zoomLevel: ZoomLevel;
  onZoomChange: (zoom: ZoomLevel) => void;
}) => {
  return (
    <Box className="sticky left-0 z-20 w-[34rem] shrink-0 border-r-[0.5px] border-gray-200/60 bg-white dark:border-dark-100 dark:bg-dark">
      <Header
        onReset={onReset}
        onZoomChange={onZoomChange}
        zoomLevel={zoomLevel}
      />
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

// Chart Component
const Chart = ({
  stories,
  dateRange,
  onDateUpdate,
  containerRef,
  zoomLevel,
  onDragStateChange,
}: {
  stories: Story[];
  dateRange: { start: Date; end: Date };
  onDateUpdate: (storyId: string, startDate: string, endDate: string) => void;
  containerRef: React.RefObject<HTMLDivElement | null>;
  zoomLevel: ZoomLevel;
  onDragStateChange: (isDragging: boolean) => void;
}) => {
  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const columnWidth = getColumnWidth(zoomLevel);

  // Calculate minimum width for timeline
  const timelineMinWidth = periods.length * columnWidth;

  // Get today's date for comparison
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  return (
    <Box className="flex-1" style={{ minWidth: `${timelineMinWidth}px` }}>
      <TimelineHeader dateRange={dateRange} zoomLevel={zoomLevel} />
      {stories.map((story) => (
        <Box className="relative h-14" key={story.id}>
          {/* Vertical grid lines for each period */}
          <Flex className="absolute inset-0">
            {periods.map((period) => {
              let isToday = false;
              let dayIsYesterday = false;

              if (zoomLevel === "weeks") {
                isToday = period.getTime() === today.getTime();
                dayIsYesterday = isYesterday(period);
              }

              return (
                <Box
                  className={cn(
                    "min-w-16 flex-1 border-r-[0.5px] border-gray-100 dark:border-dark-100",
                    {
                      "bg-gray-50/50 dark:bg-dark-200/20":
                        zoomLevel === "weeks" && isWeekend(period) && !isToday,
                      "border-primary/50 bg-primary/10 dark:border-primary/50":
                        isToday,
                      "border-primary/50 dark:border-primary/50":
                        dayIsYesterday,
                    },
                  )}
                  key={period.getTime()}
                  style={{ minWidth: `${columnWidth}px` }}
                />
              );
            })}
          </Flex>

          {/* Gantt bar on top of grid */}
          <Box className="z-5 relative h-full px-2">
            <Bar
              containerRef={containerRef}
              dateRange={dateRange}
              onDateUpdate={onDateUpdate}
              onDragStateChange={onDragStateChange}
              story={story}
              zoomLevel={zoomLevel}
            />
          </Box>
        </Box>
      ))}
    </Box>
  );
};

// Main GanttBoard Component
type GanttBoardProps = {
  stories: Story[];
  className?: string;
};

export const GanttBoard = ({ stories, className }: GanttBoardProps) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const { mutate } = useUpdateStoryMutation();
  const containerRef = useRef<HTMLDivElement>(null);
  const hasScrolledRef = useRef(false);
  const [zoomLevel, setZoomLevel] = useState<ZoomLevel>("weeks");

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

  // Update the drag state tracking
  const handleDragStateChange = useCallback((_dragging: boolean) => {
    // Simplified - no longer tracking drag state for scroll prevention
  }, []);

  // Scroll to today function - only called on mount and when Today button is clicked
  const scrollToToday = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    // Wait for next frame to ensure everything is rendered
    requestAnimationFrame(() => {
      const now = new Date();
      const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
      const columnWidth = getColumnWidth(zoomLevel);

      let periodOffset = 0;
      switch (zoomLevel) {
        case "weeks":
          periodOffset = differenceInDays(now, dateRange.start);
          break;
        case "months": {
          // Find the current month in the periods array
          const currentMonth = periods.findIndex(
            (period) =>
              period.getMonth() === now.getMonth() &&
              period.getFullYear() === now.getFullYear(),
          );
          periodOffset = currentMonth >= 0 ? currentMonth : 0;
          break;
        }
        case "quarters": {
          // Find the current quarter in the periods array
          const currentQuarter = Math.floor(now.getMonth() / 3);
          const currentQuarterIndex = periods.findIndex(
            (period) =>
              Math.floor(period.getMonth() / 3) === currentQuarter &&
              period.getFullYear() === now.getFullYear(),
          );
          periodOffset = currentQuarterIndex >= 0 ? currentQuarterIndex : 0;
          break;
        }
      }

      // Calculate position: sticky columns width + (current period position * column width)
      const stickyColumnsWidth = 34 * 16; // 34rem converted to px
      const currentPeriodPixelPosition = periodOffset * columnWidth;

      // Center the current period in the viewport
      const viewportWidth = container.clientWidth;
      const scrollPosition =
        stickyColumnsWidth + currentPeriodPixelPosition - viewportWidth / 2;

      container.scrollLeft = Math.max(0, scrollPosition);
    });
  }, [dateRange.start, zoomLevel]);

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
        <Stories
          onReset={scrollToToday}
          onZoomChange={setZoomLevel}
          stories={storiesWithDates}
          teamCode={teamCode}
          zoomLevel={zoomLevel}
        />
        <Chart
          containerRef={containerRef}
          dateRange={dateRange}
          onDateUpdate={handleDateUpdate}
          onDragStateChange={handleDragStateChange}
          stories={storiesWithDates}
          zoomLevel={zoomLevel}
        />
      </Flex>
    </div>
  );
};
