"use client";

import { format, isWeekend, isYesterday } from "date-fns";
import { cn } from "lib";
import type { MouseEvent as ReactMouseEvent, ReactNode } from "react";
import { useCallback, useMemo, useState } from "react";
import { Box, Flex, Text } from "ui";
import { BaseGanttBar } from "./base-gantt-bar";
import { BaseGanttOffscreenIndicator } from "./base-gantt-offscreen-indicator";
import { BaseGanttTimelineHeader } from "./base-gantt-timeline-header";
import type {
  GanttDateRange,
  GanttItem,
  GanttViewport,
} from "./base-gantt-types";
import {
  calculateTimelineDateFromPosition,
  calculateTimelineDatePosition,
  getColumnWidth,
  getTimePeriodsForZoom,
  type ZoomLevel,
} from "./base-gantt-utils";
import type { RenderedGanttRow } from "./base-gantt-row-window";

export const BaseGanttTimelineChart = <T extends GanttItem>({
  itemCount,
  rows,
  totalRowsHeight,
  dateRange,
  onDateUpdate,
  onBarClick,
  zoomLevel,
  renderBarContent,
  rowHeight,
  viewport,
  onScrollToItem,
  barClassName,
  onInteractionChange,
  virtualized,
}: {
  itemCount: number;
  rows: RenderedGanttRow<T>[];
  totalRowsHeight: number;
  dateRange: GanttDateRange;
  onDateUpdate: (itemId: string, startDate: string, endDate: string) => void;
  onBarClick?: (item: T) => void;
  zoomLevel: ZoomLevel;
  renderBarContent: (item: T) => ReactNode;
  rowHeight: number | string;
  viewport: GanttViewport;
  onScrollToItem: (item: T) => void;
  barClassName?: string;
  onInteractionChange: (itemId: string | null) => void;
  virtualized: boolean;
}) => {
  const [hoverPosition, setHoverPosition] = useState<number | null>(null);
  const periods = useMemo(
    () => getTimePeriodsForZoom(dateRange, zoomLevel),
    [dateRange, zoomLevel],
  );
  const columnWidth = getColumnWidth(zoomLevel);
  const timelineMinWidth = periods.length * columnWidth;

  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const todayPosition = calculateTimelineDatePosition({
    date: today,
    dateRange,
    zoomLevel,
  });
  const hoverDate =
    hoverPosition === null
      ? null
      : calculateTimelineDateFromPosition({
          position: hoverPosition,
          dateRange,
          zoomLevel,
        });

  const handleTimelineMouseMove = useCallback(
    (event: ReactMouseEvent<HTMLDivElement>) => {
      const bounds = event.currentTarget.getBoundingClientRect();
      setHoverPosition(
        Math.min(Math.max(event.clientX - bounds.left, 0), timelineMinWidth),
      );
    },
    [timelineMinWidth],
  );

  return (
    <Box
      className="relative z-0 min-h-full flex-1"
      onMouseLeave={() => {
        setHoverPosition(null);
      }}
      onMouseMove={handleTimelineMouseMove}
      style={{ minWidth: `${timelineMinWidth}px` }}
    >
      <Flex className="pointer-events-none absolute inset-x-0 top-16 bottom-0">
        {periods.map((period) => {
          const dayIsYesterday = zoomLevel === "weeks" && isYesterday(period);

          return (
            <Box
              className={cn(
                "border-border dark:border-border/40 min-w-16 flex-1 border-r-[0.5px]",
                {
                  "bg-surface-muted":
                    zoomLevel === "weeks" && isWeekend(period),
                  "border-primary/50 dark:border-primary/50": dayIsYesterday,
                },
              )}
              key={period.getTime()}
              style={{ minWidth: `${columnWidth}px` }}
            />
          );
        })}
      </Flex>

      <Box
        aria-hidden
        className="pointer-events-none absolute inset-y-0 z-20"
        style={{ left: `${todayPosition}px` }}
      >
        <Box className="bg-primary/35 absolute inset-y-0 w-px -translate-x-1/2" />
        <Text
          as="span"
          className="bg-primary text-primary-foreground absolute top-7 flex h-6 -translate-x-1/2 items-center rounded-lg px-2 text-[0.85rem] leading-none whitespace-nowrap shadow-sm"
          fontWeight="medium"
        >
          {format(today, "MMM d").toUpperCase()}
        </Text>
      </Box>

      {hoverPosition !== null && hoverDate ? (
        <Box
          aria-hidden
          className="pointer-events-none absolute inset-y-0 z-30"
          style={{ left: `${hoverPosition}px` }}
        >
          <Box className="bg-foreground/25 absolute inset-y-0 w-px -translate-x-1/2" />
          <Text
            as="span"
            className="border-border bg-surface/90 text-foreground dark:border-border/70 absolute top-7 flex h-6 -translate-x-1/2 items-center rounded-lg border-[0.5px] px-2 text-[0.85rem] leading-none whitespace-nowrap shadow-sm backdrop-blur-2xl"
            fontWeight="medium"
          >
            {format(hoverDate, "MMM d").toUpperCase()}
          </Text>
        </Box>
      ) : null}

      <Box className="relative z-1">
        <BaseGanttTimelineHeader dateRange={dateRange} zoomLevel={zoomLevel} />
        <Box
          aria-label="Timeline items"
          className="relative"
          role="list"
          style={virtualized ? { height: totalRowsHeight } : undefined}
        >
          {rows.map(({ index, item, size, start }) => (
            <Box
              aria-posinset={index + 1}
              aria-setsize={itemCount}
              className={cn(
                "border-border hover:bg-state-hover/50 dark:border-border/40 border-b-[0.5px] dark:hover:bg-white/[0.02]",
                virtualized ? "absolute inset-x-0 top-0" : "relative",
              )}
              data-gantt-item-id={item.id}
              key={item.id}
              role="listitem"
              style={
                virtualized
                  ? {
                      height: size,
                      transform: `translateY(${start}px)`,
                    }
                  : {
                      height:
                        typeof rowHeight === "number"
                          ? `${rowHeight}px`
                          : rowHeight,
                    }
              }
            >
              <Box className="relative h-full px-2">
                <BaseGanttBar
                  className={barClassName}
                  dateRange={dateRange}
                  item={item}
                  onBarClick={onBarClick}
                  onDateUpdate={onDateUpdate}
                  onInteractionChange={onInteractionChange}
                  renderContent={renderBarContent}
                  zoomLevel={zoomLevel}
                />
                <BaseGanttOffscreenIndicator
                  dateRange={dateRange}
                  item={item}
                  onScrollToItem={onScrollToItem}
                  viewport={viewport}
                  zoomLevel={zoomLevel}
                />
              </Box>
            </Box>
          ))}
        </Box>
      </Box>
    </Box>
  );
};
