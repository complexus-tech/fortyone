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
} from "date-fns";
import type { ReactNode } from "react";
import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { ArrowDown2Icon } from "icons";
import { useLocalStorage } from "@/hooks";

// Types
export type ZoomLevel = "weeks" | "months" | "quarters";

export type GanttItem = {
  id: string;
  startDate?: string | null;
  endDate?: string | null;
};

type BaseGanttProps<T extends GanttItem> = {
  items: T[];
  className?: string;
  storageKey: string; // for zoom level persistence
  onDateUpdate: (itemId: string, startDate: string, endDate: string) => void;
  renderSidebar: (
    items: T[],
    onReset: () => void,
    zoomLevel: ZoomLevel,
    onZoomChange: (zoom: ZoomLevel) => void,
  ) => ReactNode;
  renderBarContent: (item: T) => ReactNode;
};

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
      return 64;
    case "months":
      return 120;
    case "quarters":
      return 180;
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

    const isEndOfWeek =
      i === days.length - 1 ||
      !isSameWeek(currentDay, nextDay, { weekStartsOn: 0 });

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

  start.setHours(0, 0, 0, 0);
  end.setHours(0, 0, 0, 0);

  return { start, end };
};

// Generic Bar Component
const Bar = <T extends GanttItem>({
  item,
  dateRange,
  onDateUpdate,
  containerRef,
  zoomLevel,
  renderContent,
}: {
  item: T;
  dateRange: { start: Date; end: Date };
  onDateUpdate: (itemId: string, startDate: string, endDate: string) => void;
  containerRef: React.RefObject<HTMLDivElement | null>;
  zoomLevel: ZoomLevel;
  renderContent: (item: T) => ReactNode;
}) => {
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<{
    x: number;
    type: "move" | "resize-start" | "resize-end";
    originalStartDate: Date;
    originalEndDate: Date;
  } | null>(null);

  const [dragPosition, setDragPosition] = useState<{
    pixelOffsetX: number;
    originalLeft: number;
    originalWidth: number;
  } | null>(null);

  const startDate = useMemo(() => {
    const date = item.startDate ? new Date(item.startDate) : new Date();
    date.setHours(0, 0, 0, 0);
    return date;
  }, [item.startDate]);

  const endDate = useMemo(() => {
    const date = item.endDate ? new Date(item.endDate) : addDays(startDate, 1);
    date.setHours(0, 0, 0, 0);
    return date;
  }, [item.endDate, startDate]);

  const columnWidth = getColumnWidth(zoomLevel);

  const handleMouseDown = useCallback(
    (e: React.MouseEvent, type: "move" | "resize-start" | "resize-end") => {
      e.preventDefault();
      e.stopPropagation();
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

      // Get original position for reference
      const originalStartOffset = differenceInDays(
        dragStart.originalStartDate,
        dateRange.start,
      );
      const originalDuration =
        differenceInDays(
          dragStart.originalEndDate,
          dragStart.originalStartDate,
        ) || 1;

      let originalLeft: number;
      let originalWidth: number;

      // Calculate original visual position based on zoom level
      switch (zoomLevel) {
        case "weeks": {
          originalLeft = originalStartOffset * columnWidth;
          originalWidth = originalDuration * columnWidth;
          break;
        }
        case "months":
        case "quarters": {
          const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
          const totalDays = differenceInDays(dateRange.end, dateRange.start);
          const daysPerPeriod = totalDays / periods.length;

          originalLeft = (originalStartOffset / daysPerPeriod) * columnWidth;
          originalWidth = (originalDuration / daysPerPeriod) * columnWidth;
          break;
        }
        default:
          originalLeft = 0;
          originalWidth = columnWidth;
      }

      // Store raw pixel movement for 1:1 mouse responsiveness
      setDragPosition({
        pixelOffsetX: deltaX,
        originalLeft,
        originalWidth,
      });
    },
    [isDragging, dragStart, dateRange, zoomLevel, containerRef],
  );

  const handleMouseUp = useCallback(() => {
    if (!isDragging || !dragStart || !dragPosition) {
      setIsDragging(false);
      setDragStart(null);
      setDragPosition(null);
      return;
    }

    // Convert final pixel position back to day units using correct scaling for zoom level
    const columnWidth = getColumnWidth(zoomLevel);
    let finalStartDayOffset: number;
    let finalDurationDays: number;

    // Calculate conversion factor based on zoom level
    let pixelsToDayRatio: number;
    switch (zoomLevel) {
      case "weeks": {
        // Direct 1:1 ratio in weeks view
        pixelsToDayRatio = 1 / columnWidth; // 1 day per 64px
        break;
      }
      case "months":
      case "quarters": {
        // In months/quarters, need to account for the scaling
        const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
        const totalDays = differenceInDays(dateRange.end, dateRange.start);
        const daysPerPeriod = totalDays / periods.length;
        pixelsToDayRatio = daysPerPeriod / columnWidth; // days per pixel in this view
        break;
      }
      default:
        pixelsToDayRatio = 1 / 64;
    }

    if (dragStart.type === "move") {
      // For move: convert final left position to day offset
      const finalLeftPosition =
        dragPosition.originalLeft + dragPosition.pixelOffsetX;
      finalStartDayOffset = finalLeftPosition * pixelsToDayRatio;
      finalDurationDays = dragPosition.originalWidth * pixelsToDayRatio;
    } else if (dragStart.type === "resize-start") {
      // For resize-start: adjust start position, keep end the same
      const finalLeftPosition =
        dragPosition.originalLeft + dragPosition.pixelOffsetX;
      const originalEndPosition =
        dragPosition.originalLeft + dragPosition.originalWidth;
      finalStartDayOffset = finalLeftPosition * pixelsToDayRatio;
      finalDurationDays =
        (originalEndPosition - finalLeftPosition) * pixelsToDayRatio;
    } else {
      // For resize-end: keep start the same, adjust width
      finalStartDayOffset = dragPosition.originalLeft * pixelsToDayRatio;
      finalDurationDays =
        (dragPosition.originalWidth + dragPosition.pixelOffsetX) *
        pixelsToDayRatio;
    }

    // Round to nearest day for final positioning
    const roundedStartOffset = Math.round(finalStartDayOffset);
    const roundedDuration = Math.max(1, Math.round(finalDurationDays));

    const finalStartDate = addDays(dateRange.start, roundedStartOffset);
    const finalEndDate = addDays(finalStartDate, roundedDuration);

    const originalStartISO = formatISO(dragStart.originalStartDate);
    const originalEndISO = formatISO(dragStart.originalEndDate);
    const finalStartISO = formatISO(finalStartDate);
    const finalEndISO = formatISO(finalEndDate);

    if (originalStartISO !== finalStartISO || originalEndISO !== finalEndISO) {
      onDateUpdate(item.id, finalStartISO, finalEndISO);
    }

    setIsDragging(false);
    setDragStart(null);
    setDragPosition(null);
  }, [
    isDragging,
    dragStart,
    dragPosition,
    dateRange,
    zoomLevel,
    item.id,
    onDateUpdate,
  ]);

  useEffect(() => {
    if (!isDragging) return;

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);

    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isDragging, handleMouseMove, handleMouseUp]);

  const calculatePosition = () => {
    // Use drag position if dragging, otherwise calculate from actual dates
    if (dragPosition) {
      // During drag: apply pixel offset directly for 1:1 mouse movement
      let finalLeft = dragPosition.originalLeft;
      let finalWidth = dragPosition.originalWidth;

      if (dragStart?.type === "move") {
        finalLeft += dragPosition.pixelOffsetX;
      } else if (dragStart?.type === "resize-start") {
        finalLeft += dragPosition.pixelOffsetX;
        finalWidth -= dragPosition.pixelOffsetX;
      } else if (dragStart?.type === "resize-end") {
        finalWidth += dragPosition.pixelOffsetX;
      }

      return {
        leftPosition: Math.max(0, finalLeft),
        width: Math.max(10, finalWidth), // Minimum width for visibility
      };
    }

    // When not dragging: calculate from actual dates
    const startDayOffset = differenceInDays(startDate, dateRange.start);
    const durationDays = differenceInDays(endDate, startDate) || 1;

    const columnWidth = getColumnWidth(zoomLevel);

    switch (zoomLevel) {
      case "weeks": {
        return {
          leftPosition: startDayOffset * columnWidth,
          width: durationDays * columnWidth,
        };
      }
      case "months":
      case "quarters": {
        const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
        const totalDays = differenceInDays(dateRange.end, dateRange.start);
        const daysPerPeriod = totalDays / periods.length;

        return {
          leftPosition: (startDayOffset / daysPerPeriod) * columnWidth,
          width: (durationDays / daysPerPeriod) * columnWidth,
        };
      }
      default:
        return {
          leftPosition: 0,
          width: columnWidth,
        };
    }
  };

  const { leftPosition: calculatedLeft, width: calculatedWidth } =
    calculatePosition();

  const finalLeftPosition = calculatedLeft;
  const finalWidth = calculatedWidth;

  if (!item.startDate || !item.endDate) return null;
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
      <Box
        className="absolute left-0 top-0 h-full w-1 cursor-ew-resize"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-start");
        }}
      />

      <Box
        className="absolute right-0 top-0 h-full w-1 cursor-ew-resize"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-end");
        }}
      />

      <Box className="line-clamp-1 block truncate px-3 leading-10">
        {renderContent(item)}
      </Box>
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

  const timelineMinWidth = periods.length * columnWidth;

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const renderPeriodHeader = () => {
    switch (zoomLevel) {
      case "weeks":
        return (
          <>
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
        return (
          <>
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
        return (
          <>
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

// Header Component with Zoom Controls
export const GanttHeader = ({
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
      className="sticky top-0 z-10 h-16 border-b-[0.5px] border-gray-200/60 bg-white px-6 py-2.5 dark:border-dark-100 dark:bg-dark"
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
                onSelect={() => {
                  onZoomChange("weeks");
                }}
              >
                Weeks
              </Menu.Item>
              <Menu.Item
                onSelect={() => {
                  onZoomChange("months");
                }}
              >
                Months
              </Menu.Item>
              <Menu.Item
                onSelect={() => {
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

// Chart Component
const Chart = <T extends GanttItem>({
  items,
  dateRange,
  onDateUpdate,
  containerRef,
  zoomLevel,
  isContainerScrollable,
  renderBarContent,
}: {
  items: T[];
  dateRange: { start: Date; end: Date };
  onDateUpdate: (itemId: string, startDate: string, endDate: string) => void;
  containerRef: React.RefObject<HTMLDivElement | null>;
  zoomLevel: ZoomLevel;
  isContainerScrollable: boolean;
  renderBarContent: (item: T) => ReactNode;
}) => {
  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const columnWidth = getColumnWidth(zoomLevel);

  const timelineMinWidth = periods.length * columnWidth;

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  return (
    <Box className="flex-1" style={{ minWidth: `${timelineMinWidth}px` }}>
      <TimelineHeader dateRange={dateRange} zoomLevel={zoomLevel} />
      {items.map((item, idx) => (
        <Box className="relative h-14" key={item.id}>
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
                  style={{
                    minWidth: `${columnWidth}px`,
                    height:
                      idx === items.length - 1 && !isContainerScrollable
                        ? `calc(100dvh - 8rem - ${3.5 * (items.length - 0)}rem)`
                        : "100%",
                  }}
                />
              );
            })}
          </Flex>

          <Box className="z-5 relative h-full px-2">
            <Bar
              containerRef={containerRef}
              dateRange={dateRange}
              item={item}
              onDateUpdate={onDateUpdate}
              renderContent={renderBarContent}
              zoomLevel={zoomLevel}
            />
          </Box>
        </Box>
      ))}
    </Box>
  );
};

// Main BaseGantt Component
export const BaseGantt = <T extends GanttItem>({
  items,
  className,
  storageKey,
  onDateUpdate,
  renderSidebar,
  renderBarContent,
}: BaseGanttProps<T>) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const hasScrolledRef = useRef(false);
  const [zoomLevel, setZoomLevel] = useLocalStorage<ZoomLevel>(
    storageKey,
    "weeks" as ZoomLevel,
  );
  const [isContainerScrollable, setIsContainerScrollable] = useState(false);

  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const dateRange = getVisibleDateRange(today, 365);

  // Subscribe to container scrollable state changes
  useEffect(() => {
    const updateScrollableState = () => {
      if (containerRef.current) {
        const pageHeaderHeight = 64;
        const timelineHeaderHeight = 64;
        const rowHeight = 56;
        const availableHeight = window.innerHeight - pageHeaderHeight;
        const usedHeight =
          timelineHeaderHeight + itemsWithDates.length * rowHeight;

        const needsScroll = usedHeight > availableHeight;
        setIsContainerScrollable(needsScroll);
      }
    };

    updateScrollableState();
    window.addEventListener("resize", updateScrollableState);

    return () => {
      window.removeEventListener("resize", updateScrollableState);
    };
  }, [items.length]);

  const scrollToToday = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

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
          const currentMonth = periods.findIndex(
            (period) =>
              period.getMonth() === now.getMonth() &&
              period.getFullYear() === now.getFullYear(),
          );
          periodOffset = currentMonth >= 0 ? currentMonth : 0;
          break;
        }
        case "quarters": {
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

      const stickyColumnsWidth = 34 * 16;
      const currentPeriodPixelPosition = periodOffset * columnWidth;
      const viewportWidth = container.clientWidth;
      const scrollPosition =
        stickyColumnsWidth + currentPeriodPixelPosition - viewportWidth / 2;

      container.scrollLeft = Math.max(0, scrollPosition);
    });
  }, [dateRange, zoomLevel]);

  useEffect(() => {
    if (!hasScrolledRef.current) {
      scrollToToday();
      hasScrolledRef.current = true;
    }
  }, [scrollToToday]);

  const itemsWithDates = items.filter((item) => item.startDate || item.endDate);

  return (
    <div
      className={cn(
        "relative left-px overflow-x-auto overflow-y-auto",
        className,
      )}
      ref={containerRef}
    >
      <Flex className="min-h-[calc(100dvh-7.5rem)] min-w-max">
        {renderSidebar(itemsWithDates, scrollToToday, zoomLevel, setZoomLevel)}
        <Chart
          containerRef={containerRef}
          dateRange={dateRange}
          isContainerScrollable={isContainerScrollable}
          items={itemsWithDates}
          onDateUpdate={onDateUpdate}
          renderBarContent={renderBarContent}
          zoomLevel={zoomLevel}
        />
      </Flex>
    </div>
  );
};
