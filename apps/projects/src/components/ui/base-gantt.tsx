"use client";

import { differenceInDays } from "date-fns";
import { cn } from "lib";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Flex } from "ui";
import { useLocalStorage } from "@/hooks";
import { BaseGanttTimelineChart } from "./base-gantt-chart";
import type { BaseGanttProps, GanttItem } from "./base-gantt-types";
import {
  calculateGanttPosition,
  getColumnWidth,
  getGanttDateRange,
  getTimePeriodsForZoom,
  type ZoomLevel,
} from "./base-gantt-utils";
import { useGanttRowWindow } from "./base-gantt-row-window";

export { GanttControls, GanttHeader } from "./base-gantt-controls";
export type {
  BaseGanttProps,
  GanttDateRange,
  GanttItem,
  GanttViewport,
} from "./base-gantt-types";
export type { ZoomLevel } from "./base-gantt-utils";
export type { GanttRowWindow } from "./base-gantt-row-window";

const DEFAULT_STICKY_COLUMNS_WIDTH = 544;

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
  virtualizeRows = false,
  pinnedItemIds = [],
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
    scrollTop: 0,
    visibleHeight: 56 * 12,
    visibleWidth: 0,
  });
  const [activeInteractionId, setActiveInteractionId] = useState<string | null>(
    null,
  );
  const [focusedItemId, setFocusedItemId] = useState<string | null>(null);

  const today = useMemo(() => {
    const currentDate = new Date();
    currentDate.setHours(0, 0, 0, 0);
    return currentDate;
  }, []);
  const dateRange = useMemo(
    () => getGanttDateRange(today, items, zoomLevel),
    [items, today, zoomLevel],
  );
  const { renderedItems, renderedRows, rowWindow } = useGanttRowWindow({
    activeInteractionId,
    focusedItemId,
    items,
    pinnedItemIds,
    rowHeight,
    scrollTop: viewport.scrollTop,
    viewportHeight: viewport.visibleHeight,
    virtualizeRows,
  });

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
          scrollTop: container.scrollTop,
          visibleHeight: container.clientHeight,
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
      container.scrollLeft = Math.max(
        0,
        currentPeriodPixelPosition - visibleWidth / 2,
      );
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
    if (hasScrolledRef.current) return;
    scrollToTodayNow();
    hasScrolledRef.current = true;
  }, [scrollToTodayNow]);

  return (
    <div
      className={cn(
        "relative left-px overflow-x-auto overflow-y-auto",
        className,
      )}
      onFocusCapture={(event) => {
        const itemElement = (event.target as HTMLElement).closest<HTMLElement>(
          "[data-gantt-item-id]",
        );
        if (itemElement?.dataset.ganttItemId) {
          setFocusedItemId(itemElement.dataset.ganttItemId);
        }
      }}
      ref={containerRef}
    >
      <Flex className="min-h-full min-w-max">
        {renderSidebar(
          renderedItems,
          requestScrollToToday,
          zoomLevel,
          handleZoomLevelChange,
          rowWindow,
        )}
        <BaseGanttTimelineChart
          barClassName={barClassName}
          dateRange={dateRange}
          itemCount={items.length}
          onBarClick={onBarClick}
          onDateUpdate={onDateUpdate}
          onInteractionChange={setActiveInteractionId}
          onScrollToItem={scrollToItem}
          renderBarContent={renderBarContent}
          rowHeight={rowHeight}
          rows={renderedRows}
          totalRowsHeight={rowWindow.totalSize}
          viewport={viewport}
          virtualized={rowWindow.virtualized}
          zoomLevel={zoomLevel}
        />
      </Flex>
    </div>
  );
};
