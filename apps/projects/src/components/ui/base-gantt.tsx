"use client";

import { Box, Button, Flex, Menu, Text } from "ui";
import { cn } from "lib";
import {
  format,
  addDays,
  differenceInDays,
  formatISO,
  isWeekend,
  subDays,
  getWeek,
  isSameWeek,
  isYesterday,
  endOfMonth,
  startOfQuarter,
  endOfQuarter,
} from "date-fns";
import type { ReactNode } from "react";
import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { ArrowDown2Icon, ChevronLeftIcon, ChevronRightIcon } from "icons";
import { useLocalStorage } from "@/hooks";
import {
  calculateGanttPosition,
  getColumnWidth,
  getGanttOffscreenDirection,
  getTimePeriodsForZoom,
  type ZoomLevel,
} from "./base-gantt-utils";

// Types
export type { ZoomLevel } from "./base-gantt-utils";

const DEFAULT_STICKY_COLUMNS_WIDTH = 544;

export type GanttItem = {
  id: string;
  startDate?: string | null;
  endDate?: string | null;
};

type BaseGanttProps<T extends GanttItem> = {
  items: T[];
  className?: string;
  storageKey: string; // for zoom level persistence
  zoomLevel?: ZoomLevel;
  controlledZoomLevel?: ZoomLevel;
  onZoomLevelChange?: (zoom: ZoomLevel) => void;
  scrollToTodayRequest?: number;
  stickyColumnsWidth?: number;
  rowHeight?: number | string;
  barClassName?: string;
  onDateUpdate: (itemId: string, startDate: string, endDate: string) => void;
  onBarClick?: (item: T) => void;
  renderSidebar: (
    items: T[],
    onReset: () => void,
    zoomLevel: ZoomLevel,
    onZoomChange: (zoom: ZoomLevel) => void,
  ) => ReactNode;
  renderBarContent: (item: T) => ReactNode;
};

// Helper functions
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

const getDateRangeForZoom = <T extends GanttItem>(
  centerDate: Date,
  items: T[],
  zoomLevel: ZoomLevel,
) => {
  const viewportDays = zoomLevel === "weeks" ? 365 : 1460;
  const paddingDays = zoomLevel === "weeks" ? 30 : 120;
  const baseRange = getVisibleDateRange(centerDate, viewportDays);

  return items.reduce(
    (range, item) => {
      const itemStart = item.startDate ? new Date(item.startDate) : null;
      const itemEnd = item.endDate ? new Date(item.endDate) : null;

      if (itemStart && !Number.isNaN(itemStart.getTime())) {
        const paddedStart = subDays(itemStart, paddingDays);
        if (paddedStart < range.start) range.start = paddedStart;
      }

      if (itemEnd && !Number.isNaN(itemEnd.getTime())) {
        const paddedEnd = addDays(itemEnd, paddingDays);
        if (paddedEnd > range.end) range.end = paddedEnd;
      }

      return range;
    },
    { ...baseRange },
  );
};

// Generic Bar Component
const Bar = <T extends GanttItem>({
  item,
  dateRange,
  onDateUpdate,
  onBarClick,
  zoomLevel,
  renderContent,
  className,
}: {
  item: T;
  dateRange: { start: Date; end: Date };
  onDateUpdate: (itemId: string, startDate: string, endDate: string) => void;
  onBarClick?: (item: T) => void;
  zoomLevel: ZoomLevel;
  renderContent: (item: T) => ReactNode;
  className?: string;
}) => {
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<{
    x: number;
    type: "move" | "resize-start" | "resize-end";
    originalStartDate: Date;
    originalEndDate: Date;
    originalLeft: number;
    originalWidth: number;
  } | null>(null);

  const [dragPosition, setDragPosition] = useState<{
    pixelOffsetX: number;
  } | null>(null);

  // Track mouse down position for click vs drag detection
  const [mouseDownPos, setMouseDownPos] = useState<{
    x: number;
    y: number;
  } | null>(null);

  // Optimistic state to prevent flash during update
  const [optimisticDates, setOptimisticDates] = useState<{
    startDate: string;
    endDate: string;
  } | null>(null);

  const effectiveOptimisticDates = useMemo(() => {
    if (!optimisticDates || !item.startDate || !item.endDate) {
      return optimisticDates;
    }

    const propsStartISO = formatISO(new Date(item.startDate), {
      representation: "date",
    });
    const propsEndISO = formatISO(new Date(item.endDate), {
      representation: "date",
    });
    if (
      propsStartISO === optimisticDates.startDate &&
      propsEndISO === optimisticDates.endDate
    ) {
      return null;
    }

    return optimisticDates;
  }, [item.startDate, item.endDate, optimisticDates]);

  const startDate = useMemo(() => {
    const dateStr = effectiveOptimisticDates?.startDate || item.startDate;
    const date = dateStr ? new Date(dateStr) : new Date();
    date.setHours(0, 0, 0, 0);
    return date;
  }, [item.startDate, effectiveOptimisticDates?.startDate]);

  const endDate = useMemo(() => {
    const dateStr = effectiveOptimisticDates?.endDate || item.endDate;
    const date = dateStr ? new Date(dateStr) : addDays(startDate, 1);
    date.setHours(0, 0, 0, 0);
    return date;
  }, [item.endDate, effectiveOptimisticDates?.endDate, startDate]);

  const getPositionFromDates = useCallback(
    (start: Date, end: Date) =>
      calculateGanttPosition({ start, end, dateRange, zoomLevel }),
    [dateRange, zoomLevel],
  );

  const handleMouseDown = useCallback(
    (e: React.MouseEvent, type: "move" | "resize-start" | "resize-end") => {
      e.preventDefault();
      e.stopPropagation();

      // Track mouse down position for click detection
      setMouseDownPos({ x: e.clientX, y: e.clientY });

      // Clear any existing optimistic state when starting a new drag
      setOptimisticDates(null);

      // Capture the current visual position
      const currentPosition = getPositionFromDates(startDate, endDate);

      setIsDragging(true);
      setDragStart({
        x: e.clientX,
        type,
        originalStartDate: startDate,
        originalEndDate: endDate,
        originalLeft: currentPosition.leftPosition,
        originalWidth: currentPosition.width,
      });
    },
    [startDate, endDate, getPositionFromDates],
  );

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !dragStart) return;

      const deltaX = e.clientX - dragStart.x;

      // Store just the pixel offset for 1:1 mouse responsiveness
      setDragPosition({
        pixelOffsetX: deltaX,
      });
    },
    [isDragging, dragStart],
  );

  const handleMouseUp = useCallback(() => {
    if (!isDragging || !dragStart || !dragPosition) {
      setIsDragging(false);
      setDragStart(null);
      setDragPosition(null);
      setMouseDownPos(null);
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
        dragStart.originalLeft + dragPosition.pixelOffsetX;
      finalStartDayOffset = finalLeftPosition * pixelsToDayRatio;
      finalDurationDays = dragStart.originalWidth * pixelsToDayRatio;
    } else if (dragStart.type === "resize-start") {
      // For resize-start: adjust start position, keep end the same
      const finalLeftPosition =
        dragStart.originalLeft + dragPosition.pixelOffsetX;
      const originalEndPosition =
        dragStart.originalLeft + dragStart.originalWidth;
      finalStartDayOffset = finalLeftPosition * pixelsToDayRatio;
      finalDurationDays =
        (originalEndPosition - finalLeftPosition) * pixelsToDayRatio;
    } else {
      // For resize-end: keep start the same, adjust width
      finalStartDayOffset = dragStart.originalLeft * pixelsToDayRatio;
      finalDurationDays =
        (dragStart.originalWidth + dragPosition.pixelOffsetX) *
        pixelsToDayRatio;
    }

    // Round to nearest day for final positioning
    const roundedStartOffset = Math.round(finalStartDayOffset);
    const roundedDuration = Math.max(1, Math.round(finalDurationDays));

    const finalStartDate = addDays(dateRange.start, roundedStartOffset);
    const finalEndDate = addDays(finalStartDate, roundedDuration);

    const originalStartISO = formatISO(dragStart.originalStartDate, {
      representation: "date",
    });
    const originalEndISO = formatISO(dragStart.originalEndDate, {
      representation: "date",
    });
    const finalStartISO = formatISO(finalStartDate, {
      representation: "date",
    });
    const finalEndISO = formatISO(finalEndDate, {
      representation: "date",
    });

    if (originalStartISO !== finalStartISO || originalEndISO !== finalEndISO) {
      // Set optimistic state to maintain visual position until props update
      setOptimisticDates({
        startDate: finalStartISO,
        endDate: finalEndISO,
      });

      onDateUpdate(item.id, finalStartISO, finalEndISO);
    }

    setIsDragging(false);
    setDragStart(null);
    setDragPosition(null);
    setMouseDownPos(null);
  }, [
    isDragging,
    dragStart,
    dragPosition,
    dateRange,
    zoomLevel,
    item.id,
    onDateUpdate,
  ]);

  // Handle click detection
  const handleClick = useCallback(
    (e: React.MouseEvent) => {
      if (!onBarClick || !mouseDownPos) return;

      // Check if this was a click (not a drag)
      const deltaX = Math.abs(e.clientX - mouseDownPos.x);
      const deltaY = Math.abs(e.clientY - mouseDownPos.y);
      const clickThreshold = 5; // pixels

      if (deltaX <= clickThreshold && deltaY <= clickThreshold) {
        onBarClick(item);
      }

      setMouseDownPos(null);
    },
    [onBarClick, mouseDownPos, item],
  );

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
    if (dragPosition && dragStart) {
      // During drag: apply the same rounding logic as final calculation to prevent jumps
      const columnWidth = getColumnWidth(zoomLevel);

      // Calculate conversion factor based on zoom level (same as in handleMouseUp)
      let pixelsToDayRatio: number;
      switch (zoomLevel) {
        case "weeks": {
          pixelsToDayRatio = 1 / columnWidth;
          break;
        }
        case "months":
        case "quarters": {
          const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
          const totalDays = differenceInDays(dateRange.end, dateRange.start);
          const daysPerPeriod = totalDays / periods.length;
          pixelsToDayRatio = daysPerPeriod / columnWidth;
          break;
        }
        default:
          pixelsToDayRatio = 1 / 64;
      }

      let finalStartDayOffset: number;
      let finalDurationDays: number;

      if (dragStart.type === "move") {
        const finalLeftPosition =
          dragStart.originalLeft + dragPosition.pixelOffsetX;
        finalStartDayOffset = finalLeftPosition * pixelsToDayRatio;
        finalDurationDays = dragStart.originalWidth * pixelsToDayRatio;
      } else if (dragStart.type === "resize-start") {
        const finalLeftPosition =
          dragStart.originalLeft + dragPosition.pixelOffsetX;
        const originalEndPosition =
          dragStart.originalLeft + dragStart.originalWidth;
        finalStartDayOffset = finalLeftPosition * pixelsToDayRatio;
        finalDurationDays =
          (originalEndPosition - finalLeftPosition) * pixelsToDayRatio;
      } else {
        // resize-end
        finalStartDayOffset = dragStart.originalLeft * pixelsToDayRatio;
        finalDurationDays =
          (dragStart.originalWidth + dragPosition.pixelOffsetX) *
          pixelsToDayRatio;
      }

      // Apply the same rounding as in handleMouseUp to prevent optimistic update jumps
      const roundedStartOffset = Math.round(finalStartDayOffset);
      const roundedDuration = Math.max(1, Math.round(finalDurationDays));

      // Convert back to visual position using rounded values
      switch (zoomLevel) {
        case "weeks": {
          return {
            leftPosition: Math.max(0, roundedStartOffset * columnWidth),
            width: roundedDuration * columnWidth,
          };
        }
        case "months":
        case "quarters": {
          const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
          const totalDays = differenceInDays(dateRange.end, dateRange.start);
          const daysPerPeriod = totalDays / periods.length;

          return {
            leftPosition: Math.max(
              0,
              (roundedStartOffset / daysPerPeriod) * columnWidth,
            ),
            width: (roundedDuration / daysPerPeriod) * columnWidth,
          };
        }
        default:
          return {
            leftPosition: Math.max(0, roundedStartOffset * columnWidth),
            width: Math.max(10, roundedDuration * columnWidth),
          };
      }
    }

    // When not dragging: calculate from actual dates
    return getPositionFromDates(startDate, endDate);
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
        "group border-border/70 focus-visible:ring-primary dark:border-border dark:bg-surface/80 absolute z-0 h-10 rounded-xl border-[0.5px] bg-white/80 backdrop-blur-2xl transition-colors focus-visible:ring-1 focus-visible:outline-none",
        {
          "shadow-lg": isDragging,
          "hover:border-border-strong dark:hover:bg-surface/90 cursor-pointer hover:bg-white/90":
            onBarClick,
        },
        className,
      )}
      onKeyDown={(e) => {
        if (!onBarClick || (e.key !== "Enter" && e.key !== " ")) return;
        e.preventDefault();
        onBarClick(item);
      }}
      onMouseDown={(e) => {
        handleMouseDown(e, "move");
      }}
      onMouseUp={handleClick}
      role={onBarClick ? "button" : undefined}
      style={{
        left: `${finalLeftPosition}px`,
        width: `${finalWidth}px`,
        top: "6px",
      }}
      tabIndex={onBarClick ? 0 : -1}
    >
      <Box
        className="group-hover:bg-foreground/20 absolute top-1/2 bottom-1/2 -left-1 h-[70%] w-2 -translate-y-1/2 cursor-col-resize rounded transition-colors dark:group-hover:bg-white/25"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-start");
        }}
      />

      <Box
        className="group-hover:bg-foreground/20 absolute top-1/2 -right-1 bottom-1/2 h-[70%] w-2 -translate-y-1/2 cursor-col-resize rounded transition-colors dark:group-hover:bg-white/25"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-end");
        }}
      />
      <Box className="absolute inset-0 overflow-hidden">
        <Box className="px-3 leading-10">{renderContent(item)}</Box>
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
            <Box className="border-border/45 border-b-[0.5px]">
              <Flex>
                {getWeekSpans(periods).map(
                  ({ week, month, span, startIndex }) => (
                    <Box
                      className="border-border/45 border-r-[0.5px] px-2 py-1.5 text-left"
                      key={`${month}-${week}-${startIndex}`}
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
                  ),
                )}
              </Flex>
            </Box>

            <Flex>
              {periods.map((day) => {
                const isToday = day.getTime() === today.getTime();

                return (
                  <Box
                    className={cn(
                      "border-border/45 h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] px-1 py-1 text-center",
                      {
                        "bg-surface-muted": isWeekend(day) && !isToday,
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
            <Box className="border-border/45 border-b-[0.5px]">
              <Flex>
                {periods.map((month) => (
                  <Box
                    className="border-border/45 border-r-[0.5px] px-2 py-1.5 text-left"
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
                  className="border-border/45 h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] px-1 py-1 text-center"
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
            <Box className="border-border/45 border-b-[0.5px]">
              <Flex>
                {periods.map((quarter) => (
                  <Box
                    className="border-border/45 border-r-[0.5px] px-2 py-1.5 text-left"
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
                    className="border-border/45 h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] px-1 py-1 text-center"
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
      className="border-border/45 bg-background sticky top-0 z-10 h-16 border-b-[0.5px]"
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
  children,
}: {
  onReset: () => void;
  zoomLevel: ZoomLevel;
  onZoomChange: (zoom: ZoomLevel) => void;
  children?: ReactNode;
}) => {
  return (
    <Box className="border-border bg-background sticky top-0 z-10 hidden h-16 border-b-[0.5px] px-4 md:block">
      <GanttControls
        className={children ? "h-9" : "h-full"}
        onReset={onReset}
        onZoomChange={onZoomChange}
        zoomLevel={zoomLevel}
      />
      {children}
    </Box>
  );
};

export const GanttControls = ({
  onReset,
  zoomLevel,
  onZoomChange,
  className,
  showSeparator = false,
}: {
  onReset: () => void;
  zoomLevel: ZoomLevel;
  onZoomChange: (zoom: ZoomLevel) => void;
  className?: string;
  showSeparator?: boolean;
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
    <Flex align="center" className={className} gap={2}>
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
      {showSeparator ? (
        <span className="text-text-secondary mx-1 hidden opacity-40 md:inline">
          |
        </span>
      ) : null}
      <Button color="tertiary" onClick={onReset} size="sm">
        Today
      </Button>
    </Flex>
  );
};

// Chart Component
const Chart = <T extends GanttItem>({
  items,
  dateRange,
  onDateUpdate,
  onBarClick,
  zoomLevel,
  renderBarContent,
  rowHeight,
  viewport,
  onScrollToItem,
  barClassName,
}: {
  items: T[];
  dateRange: { start: Date; end: Date };
  onDateUpdate: (itemId: string, startDate: string, endDate: string) => void;
  onBarClick?: (item: T) => void;
  zoomLevel: ZoomLevel;
  renderBarContent: (item: T) => ReactNode;
  rowHeight: number | string;
  viewport: { scrollLeft: number; visibleWidth: number };
  onScrollToItem: (item: T) => void;
  barClassName?: string;
}) => {
  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const columnWidth = getColumnWidth(zoomLevel);

  const timelineMinWidth = periods.length * columnWidth;

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  return (
    <Box
      className="relative min-h-full flex-1"
      style={{ minWidth: `${timelineMinWidth}px` }}
    >
      <Flex className="pointer-events-none absolute inset-x-0 top-16 bottom-0">
        {periods.map((period) => {
          const isToday =
            zoomLevel === "weeks" && period.getTime() === today.getTime();
          const dayIsYesterday = zoomLevel === "weeks" && isYesterday(period);

          return (
            <Box
              className={cn(
                "border-border/40 min-w-16 flex-1 border-r-[0.5px]",
                {
                  "bg-surface-muted":
                    zoomLevel === "weeks" && isWeekend(period) && !isToday,
                  "border-primary/50 bg-primary/10 dark:border-primary/50":
                    isToday,
                  "border-primary/50 dark:border-primary/50": dayIsYesterday,
                },
              )}
              key={period.getTime()}
              style={{ minWidth: `${columnWidth}px` }}
            />
          );
        })}
      </Flex>

      <Box className="relative z-1">
        <TimelineHeader dateRange={dateRange} zoomLevel={zoomLevel} />
        {items.map((item) => (
          <Box
            className="border-border/40 hover:bg-state-hover relative border-b-[0.5px] dark:hover:bg-white/[0.02]"
            key={item.id}
            style={{
              height:
                typeof rowHeight === "number" ? `${rowHeight}px` : rowHeight,
            }}
          >
            <Box className="relative h-full px-2">
              <Bar
                className={barClassName}
                dateRange={dateRange}
                item={item}
                onBarClick={onBarClick}
                onDateUpdate={onDateUpdate}
                renderContent={renderBarContent}
                zoomLevel={zoomLevel}
              />
              {(() => {
                if (!item.startDate || !item.endDate) return null;

                const position = calculateGanttPosition({
                  start: new Date(item.startDate),
                  end: new Date(item.endDate),
                  dateRange,
                  zoomLevel,
                });
                const direction = getGanttOffscreenDirection(
                  position,
                  viewport,
                );

                if (!direction || viewport.visibleWidth <= 0) return null;

                const indicatorPosition =
                  direction === "left"
                    ? viewport.scrollLeft + 12
                    : viewport.scrollLeft + viewport.visibleWidth - 12;
                const dateRangeLabel = `${format(
                  new Date(item.startDate),
                  "MMM yyyy",
                )} - ${format(new Date(item.endDate), "MMM yyyy")}`;

                return (
                  <Flex
                    align="center"
                    className="absolute top-3 z-20 gap-2"
                    style={{
                      left: indicatorPosition,
                      transform:
                        direction === "right" ? "translateX(-100%)" : undefined,
                    }}
                  >
                    {direction === "right" ? (
                      <Text
                        className="shrink-0 text-[0.95rem] whitespace-nowrap opacity-70"
                        color="muted"
                      >
                        {dateRangeLabel}
                      </Text>
                    ) : null}
                    <Button
                      aria-label={`Scroll ${direction} to item`}
                      asIcon
                      className="border-border/70 dark:border-border dark:bg-surface/80 h-8 w-8 bg-white/80 p-0 backdrop-blur-2xl"
                      color="tertiary"
                      leftIcon={
                        direction === "left" ? (
                          <ChevronLeftIcon className="h-3.5 w-auto" />
                        ) : (
                          <ChevronRightIcon className="h-3.5 w-auto" />
                        )
                      }
                      onClick={() => {
                        onScrollToItem(item);
                      }}
                      rounded="full"
                      size="sm"
                      variant="outline"
                    />
                    {direction === "left" ? (
                      <Text
                        className="shrink-0 text-[0.95rem] whitespace-nowrap opacity-70"
                        color="muted"
                      >
                        {dateRangeLabel}
                      </Text>
                    ) : null}
                  </Flex>
                );
              })()}
            </Box>
          </Box>
        ))}
      </Box>
    </Box>
  );
};

// Main BaseGantt Component
export const BaseGantt = <T extends GanttItem>({
  items,
  className,
  storageKey,
  onDateUpdate,
  onBarClick,
  renderSidebar,
  renderBarContent,
  zoomLevel: defaultZoomLevel = "weeks",
  controlledZoomLevel,
  onZoomLevelChange,
  scrollToTodayRequest = 0,
  stickyColumnsWidth = DEFAULT_STICKY_COLUMNS_WIDTH,
  rowHeight = "3.5rem",
  barClassName,
}: BaseGanttProps<T>) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const hasScrolledRef = useRef(false);
  const [storedZoomLevel, setStoredZoomLevel] = useLocalStorage<ZoomLevel>(
    storageKey,
    defaultZoomLevel,
  );
  const zoomLevel = controlledZoomLevel ?? storedZoomLevel;
  const [scrollToTodayRequests, setScrollToTodayRequests] = useState(0);
  const [viewport, setViewport] = useState({
    scrollLeft: 0,
    visibleWidth: 0,
  });

  const today = useMemo(() => {
    const currentDate = new Date();
    currentDate.setHours(0, 0, 0, 0);
    return currentDate;
  }, []);
  const dateRange = useMemo(
    () => getDateRangeForZoom(today, items, zoomLevel),
    [items, today, zoomLevel],
  );

  const getRenderedStickyColumnsWidth = useCallback(() => {
    const container = containerRef.current;
    const stickyPane = container?.firstElementChild
      ?.firstElementChild as HTMLElement | null;
    const renderedWidth = stickyPane?.getBoundingClientRect().width;

    return renderedWidth && renderedWidth > 0
      ? renderedWidth
      : stickyColumnsWidth;
  }, [stickyColumnsWidth]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let frameId: number | undefined;
    const updateViewport = () => {
      if (frameId !== undefined) cancelAnimationFrame(frameId);
      frameId = requestAnimationFrame(() => {
        const renderedStickyColumnsWidth = getRenderedStickyColumnsWidth();
        setViewport({
          scrollLeft: container.scrollLeft,
          visibleWidth: Math.max(
            0,
            container.clientWidth - renderedStickyColumnsWidth,
          ),
        });
      });
    };

    updateViewport();
    container.addEventListener("scroll", updateViewport, { passive: true });
    window.addEventListener("resize", updateViewport);

    return () => {
      if (frameId !== undefined) cancelAnimationFrame(frameId);
      container.removeEventListener("scroll", updateViewport);
      window.removeEventListener("resize", updateViewport);
    };
  }, [getRenderedStickyColumnsWidth]);

  const requestScrollToToday = useCallback(() => {
    setScrollToTodayRequests((current) => current + 1);
  }, []);

  const handleZoomLevelChange = useCallback(
    (nextZoomLevel: ZoomLevel) => {
      if (controlledZoomLevel === undefined) {
        setStoredZoomLevel(nextZoomLevel);
      }
      onZoomLevelChange?.(nextZoomLevel);
    },
    [controlledZoomLevel, onZoomLevelChange, setStoredZoomLevel],
  );

  const scrollToTodayNow = useCallback(() => {
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

      const currentPeriodPixelPosition = periodOffset * columnWidth;
      const visibleWidth = Math.max(
        0,
        container.clientWidth - getRenderedStickyColumnsWidth(),
      );
      const scrollPosition = currentPeriodPixelPosition - visibleWidth / 2;

      container.scrollLeft = Math.max(0, scrollPosition);
    });
  }, [dateRange, getRenderedStickyColumnsWidth, zoomLevel]);

  const scrollToItem = useCallback(
    (item: T) => {
      const container = containerRef.current;
      if (!container || !item.startDate || !item.endDate) return;

      const position = calculateGanttPosition({
        start: new Date(item.startDate),
        end: new Date(item.endDate),
        dateRange,
        zoomLevel,
      });
      const visibleWidth = Math.max(
        0,
        container.clientWidth - getRenderedStickyColumnsWidth(),
      );
      const itemCenter = position.leftPosition + position.width / 2;

      container.scrollTo({
        behavior: "smooth",
        left: Math.max(0, itemCenter - visibleWidth / 2),
      });
    },
    [dateRange, getRenderedStickyColumnsWidth, zoomLevel],
  );

  useEffect(() => {
    if (scrollToTodayRequests + scrollToTodayRequest === 0) return;
    scrollToTodayNow();
  }, [scrollToTodayNow, scrollToTodayRequest, scrollToTodayRequests]);

  useEffect(() => {
    if (!hasScrolledRef.current) {
      scrollToTodayNow();
      hasScrolledRef.current = true;
    }
  }, [scrollToTodayNow]);

  return (
    <div
      className={cn(
        "relative left-px overflow-x-auto overflow-y-auto",
        className,
      )}
      ref={containerRef}
    >
      <Flex className="min-h-full min-w-max">
        {renderSidebar(
          items,
          requestScrollToToday,
          zoomLevel,
          handleZoomLevelChange,
        )}
        <Chart
          barClassName={barClassName}
          dateRange={dateRange}
          items={items}
          onBarClick={onBarClick}
          onDateUpdate={onDateUpdate}
          onScrollToItem={scrollToItem}
          renderBarContent={renderBarContent}
          rowHeight={rowHeight}
          viewport={viewport}
          zoomLevel={zoomLevel}
        />
      </Flex>
    </div>
  );
};
