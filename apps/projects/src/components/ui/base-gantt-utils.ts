import {
  differenceInDays,
  eachDayOfInterval,
  eachMonthOfInterval,
  eachQuarterOfInterval,
  endOfMonth,
  endOfQuarter,
  startOfMonth,
  startOfQuarter,
} from "date-fns";

export type ZoomLevel = "weeks" | "months" | "quarters";
export type OffscreenDirection = "left" | "right";

export const getTimePeriodsForZoom = (
  dateRange: { start: Date; end: Date },
  zoomLevel: ZoomLevel,
) => {
  switch (zoomLevel) {
    case "weeks":
      return eachDayOfInterval(dateRange);
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
  }
};

export const getColumnWidth = (zoomLevel: ZoomLevel) => {
  switch (zoomLevel) {
    case "weeks":
      return 64;
    case "months":
      return 120;
    case "quarters":
      return 180;
  }
};

export const calculateGanttPosition = ({
  start,
  end,
  dateRange,
  zoomLevel,
}: {
  start: Date;
  end: Date;
  dateRange: { start: Date; end: Date };
  zoomLevel: ZoomLevel;
}) => {
  const columnWidth = getColumnWidth(zoomLevel);
  const startDayOffset = differenceInDays(start, dateRange.start);
  const durationDays = Math.max(1, differenceInDays(end, start));

  if (zoomLevel === "weeks") {
    return {
      leftPosition: startDayOffset * columnWidth,
      width: durationDays * columnWidth,
    };
  }

  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const totalDays = differenceInDays(dateRange.end, dateRange.start);
  const daysPerPeriod = totalDays / periods.length;

  return {
    leftPosition: (startDayOffset / daysPerPeriod) * columnWidth,
    width: (durationDays / daysPerPeriod) * columnWidth,
  };
};

export const getGanttOffscreenDirection = (
  position: { leftPosition: number; width: number },
  viewport: { scrollLeft: number; visibleWidth: number },
): OffscreenDirection | null => {
  if (position.leftPosition + position.width < viewport.scrollLeft) {
    return "left";
  }

  if (position.leftPosition > viewport.scrollLeft + viewport.visibleWidth) {
    return "right";
  }

  return null;
};
