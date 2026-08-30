"use client";

import { formatISO } from "date-fns";
import { cn } from "lib";
import type { MouseEvent as ReactMouseEvent, ReactNode } from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Box } from "ui";
import type { GanttDateRange, GanttItem } from "./base-gantt-types";
import {
  getGanttBarDatesForDrag,
  getGanttBarDragPosition,
  type GanttBarDragStart,
  type GanttBarInteractionType,
} from "./base-gantt-bar-utils";
import { calculateGanttPosition, type ZoomLevel } from "./base-gantt-utils";

export const BaseGanttBar = <T extends GanttItem>({
  item,
  dateRange,
  onDateUpdate,
  onBarClick,
  zoomLevel,
  renderContent,
  className,
  onInteractionChange,
}: {
  item: T;
  dateRange: GanttDateRange;
  onDateUpdate: (itemId: string, startDate: string, endDate: string) => void;
  onBarClick?: (item: T) => void;
  zoomLevel: ZoomLevel;
  renderContent: (item: T) => ReactNode;
  className?: string;
  onInteractionChange?: (itemId: string | null) => void;
}) => {
  const interactionActiveRef = useRef(false);
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<GanttBarDragStart | null>(null);
  const [dragPosition, setDragPosition] = useState<{
    pixelOffsetX: number;
  } | null>(null);
  const [mouseDownPos, setMouseDownPos] = useState<{
    x: number;
    y: number;
  } | null>(null);
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

    return propsStartISO === optimisticDates.startDate &&
      propsEndISO === optimisticDates.endDate
      ? null
      : optimisticDates;
  }, [item.endDate, item.startDate, optimisticDates]);

  const startDate = useMemo(() => {
    const dateStr = effectiveOptimisticDates?.startDate || item.startDate;
    const date = dateStr ? new Date(dateStr) : new Date();
    date.setHours(0, 0, 0, 0);
    return date;
  }, [effectiveOptimisticDates?.startDate, item.startDate]);

  const endDate = useMemo(() => {
    const dateStr = effectiveOptimisticDates?.endDate || item.endDate;
    const date = dateStr ? new Date(dateStr) : new Date(startDate);
    if (!dateStr) date.setDate(date.getDate() + 1);
    date.setHours(0, 0, 0, 0);
    return date;
  }, [effectiveOptimisticDates?.endDate, item.endDate, startDate]);

  const getPositionFromDates = useCallback(
    (start: Date, end: Date) =>
      calculateGanttPosition({ start, end, dateRange, zoomLevel }),
    [dateRange, zoomLevel],
  );

  const resetInteraction = useCallback(() => {
    setIsDragging(false);
    setDragStart(null);
    setDragPosition(null);
    setMouseDownPos(null);
    interactionActiveRef.current = false;
    onInteractionChange?.(null);
  }, [onInteractionChange]);

  const handleMouseDown = useCallback(
    (event: ReactMouseEvent, type: GanttBarInteractionType) => {
      event.preventDefault();
      event.stopPropagation();

      setMouseDownPos({ x: event.clientX, y: event.clientY });
      interactionActiveRef.current = true;
      onInteractionChange?.(item.id);
      setOptimisticDates(null);

      const currentPosition = getPositionFromDates(startDate, endDate);
      setIsDragging(true);
      setDragStart({
        x: event.clientX,
        type,
        originalStartDate: startDate,
        originalEndDate: endDate,
        originalLeft: currentPosition.leftPosition,
        originalWidth: currentPosition.width,
      });
    },
    [endDate, getPositionFromDates, item.id, onInteractionChange, startDate],
  );

  const handleMouseMove = useCallback(
    (event: MouseEvent) => {
      if (!isDragging || !dragStart) return;

      setDragPosition({ pixelOffsetX: event.clientX - dragStart.x });
    },
    [dragStart, isDragging],
  );

  const handleMouseUp = useCallback(() => {
    if (!isDragging || !dragStart || !dragPosition) {
      resetInteraction();
      return;
    }

    const { endDate: finalEndDate, startDate: finalStartDate } =
      getGanttBarDatesForDrag({
        dateRange,
        dragStart,
        pixelOffsetX: dragPosition.pixelOffsetX,
        zoomLevel,
      });
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
      setOptimisticDates({
        startDate: finalStartISO,
        endDate: finalEndISO,
      });
      onDateUpdate(item.id, finalStartISO, finalEndISO);
    }

    resetInteraction();
  }, [
    dateRange,
    dragPosition,
    dragStart,
    isDragging,
    item.id,
    onDateUpdate,
    resetInteraction,
    zoomLevel,
  ]);

  const handleClick = useCallback(
    (event: ReactMouseEvent) => {
      if (!onBarClick || !mouseDownPos) return;

      const isClick =
        Math.abs(event.clientX - mouseDownPos.x) <= 5 &&
        Math.abs(event.clientY - mouseDownPos.y) <= 5;
      if (isClick) onBarClick(item);

      setMouseDownPos(null);
    },
    [item, mouseDownPos, onBarClick],
  );

  useEffect(() => {
    if (!isDragging) return;

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);

    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [handleMouseMove, handleMouseUp, isDragging]);

  useEffect(
    () => () => {
      if (interactionActiveRef.current) onInteractionChange?.(null);
    },
    [onInteractionChange],
  );

  const position =
    dragPosition && dragStart
      ? getGanttBarDragPosition({
          dateRange,
          dragStart,
          pixelOffsetX: dragPosition.pixelOffsetX,
          zoomLevel,
        })
      : getPositionFromDates(startDate, endDate);

  if (!item.startDate || !item.endDate || position.width <= 0) return null;

  return (
    <Box
      className={cn(
        "group border-border focus-visible:ring-primary dark:border-border/70 dark:bg-surface/80 bg-surface-muted/80 absolute z-0 h-10 rounded-xl border-[0.5px] backdrop-blur-2xl transition-colors focus-visible:ring-1 focus-visible:outline-none",
        {
          "shadow-lg": isDragging,
          "hover:border-border-strong hover:bg-surface-muted dark:hover:bg-surface/90 cursor-pointer":
            onBarClick,
        },
        className,
      )}
      onKeyDown={(event) => {
        if (!onBarClick || (event.key !== "Enter" && event.key !== " ")) {
          return;
        }
        event.preventDefault();
        onBarClick(item);
      }}
      onMouseDown={(event) => {
        handleMouseDown(event, "move");
      }}
      onMouseUp={handleClick}
      role={onBarClick ? "button" : undefined}
      style={{
        left: `${position.leftPosition}px`,
        top: "6px",
        width: `${position.width}px`,
      }}
      tabIndex={onBarClick ? 0 : -1}
    >
      <Box
        className="group-hover:bg-foreground/20 absolute top-1/2 bottom-1/2 -left-1 h-[70%] w-2 -translate-y-1/2 cursor-col-resize rounded transition-colors dark:group-hover:bg-white/25"
        onMouseDown={(event) => {
          event.stopPropagation();
          handleMouseDown(event, "resize-start");
        }}
      />
      <Box
        className="group-hover:bg-foreground/20 absolute top-1/2 -right-1 bottom-1/2 h-[70%] w-2 -translate-y-1/2 cursor-col-resize rounded transition-colors dark:group-hover:bg-white/25"
        onMouseDown={(event) => {
          event.stopPropagation();
          handleMouseDown(event, "resize-end");
        }}
      />
      <Box className="absolute inset-0 overflow-hidden rounded-[inherit]">
        <Box className="px-3 leading-10">{renderContent(item)}</Box>
      </Box>
    </Box>
  );
};
