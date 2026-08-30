import { addDays, differenceInDays } from "date-fns";
import type { GanttDateRange } from "./base-gantt-types";
import {
  getColumnWidth,
  getTimePeriodsForZoom,
  type ZoomLevel,
} from "./base-gantt-utils";

export type GanttBarInteractionType = "move" | "resize-start" | "resize-end";

export type GanttBarDragStart = {
  x: number;
  type: GanttBarInteractionType;
  originalStartDate: Date;
  originalEndDate: Date;
  originalLeft: number;
  originalWidth: number;
};

type GanttBarDragMetrics = {
  startDayOffset: number;
  durationDays: number;
};

const getPixelsToDayRatio = (
  dateRange: GanttDateRange,
  zoomLevel: ZoomLevel,
) => {
  const columnWidth = getColumnWidth(zoomLevel);
  if (zoomLevel === "weeks") return 1 / columnWidth;

  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const totalDays = differenceInDays(dateRange.end, dateRange.start);
  return totalDays / periods.length / columnWidth;
};

const getDragMetrics = ({
  dateRange,
  dragStart,
  pixelOffsetX,
  zoomLevel,
}: {
  dateRange: GanttDateRange;
  dragStart: GanttBarDragStart;
  pixelOffsetX: number;
  zoomLevel: ZoomLevel;
}): GanttBarDragMetrics => {
  const pixelsToDayRatio = getPixelsToDayRatio(dateRange, zoomLevel);

  switch (dragStart.type) {
    case "move":
      return {
        startDayOffset:
          (dragStart.originalLeft + pixelOffsetX) * pixelsToDayRatio,
        durationDays: dragStart.originalWidth * pixelsToDayRatio,
      };
    case "resize-start": {
      const finalLeftPosition = dragStart.originalLeft + pixelOffsetX;
      const originalEndPosition =
        dragStart.originalLeft + dragStart.originalWidth;

      return {
        startDayOffset: finalLeftPosition * pixelsToDayRatio,
        durationDays:
          (originalEndPosition - finalLeftPosition) * pixelsToDayRatio,
      };
    }
    case "resize-end":
      return {
        startDayOffset: dragStart.originalLeft * pixelsToDayRatio,
        durationDays:
          (dragStart.originalWidth + pixelOffsetX) * pixelsToDayRatio,
      };
  }
};

const getRoundedDragMetrics = (args: {
  dateRange: GanttDateRange;
  dragStart: GanttBarDragStart;
  pixelOffsetX: number;
  zoomLevel: ZoomLevel;
}) => {
  const { durationDays, startDayOffset } = getDragMetrics(args);

  return {
    durationDays: Math.max(1, Math.round(durationDays)),
    startDayOffset: Math.round(startDayOffset),
  };
};

export const getGanttBarDatesForDrag = ({
  dateRange,
  dragStart,
  pixelOffsetX,
  zoomLevel,
}: {
  dateRange: GanttDateRange;
  dragStart: GanttBarDragStart;
  pixelOffsetX: number;
  zoomLevel: ZoomLevel;
}) => {
  const { durationDays, startDayOffset } = getRoundedDragMetrics({
    dateRange,
    dragStart,
    pixelOffsetX,
    zoomLevel,
  });
  const startDate = addDays(dateRange.start, startDayOffset);

  return {
    endDate: addDays(startDate, durationDays),
    startDate,
  };
};

export const getGanttBarDragPosition = ({
  dateRange,
  dragStart,
  pixelOffsetX,
  zoomLevel,
}: {
  dateRange: GanttDateRange;
  dragStart: GanttBarDragStart;
  pixelOffsetX: number;
  zoomLevel: ZoomLevel;
}) => {
  const { durationDays, startDayOffset } = getRoundedDragMetrics({
    dateRange,
    dragStart,
    pixelOffsetX,
    zoomLevel,
  });
  const columnWidth = getColumnWidth(zoomLevel);

  if (zoomLevel === "weeks") {
    return {
      leftPosition: Math.max(0, startDayOffset * columnWidth),
      width: durationDays * columnWidth,
    };
  }

  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const totalDays = differenceInDays(dateRange.end, dateRange.start);
  const daysPerPeriod = totalDays / periods.length;

  return {
    leftPosition: Math.max(0, (startDayOffset / daysPerPeriod) * columnWidth),
    width: (durationDays / daysPerPeriod) * columnWidth,
  };
};
