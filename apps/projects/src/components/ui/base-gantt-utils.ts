import {
  addDays,
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

export const calculateTimelineDatePosition = ({
  date,
  dateRange,
  zoomLevel,
}: {
  date: Date;
  dateRange: { start: Date; end: Date };
  zoomLevel: ZoomLevel;
}) => {
  const columnWidth = getColumnWidth(zoomLevel);

  if (zoomLevel === "weeks") {
    return differenceInDays(date, dateRange.start) * columnWidth;
  }

  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const periodStart =
    zoomLevel === "months" ? startOfMonth(date) : startOfQuarter(date);
  const periodEnd =
    zoomLevel === "months" ? endOfMonth(date) : endOfQuarter(date);
  const periodIndex = periods.findIndex(
    (period) => period.getTime() === periodStart.getTime(),
  );

  if (periodIndex < 0) return 0;

  const daysInPeriod = differenceInDays(periodEnd, periodStart) + 1;
  const elapsedDays = differenceInDays(date, periodStart);

  return (periodIndex + elapsedDays / daysInPeriod) * columnWidth;
};

export const calculateTimelineDateFromPosition = ({
  position,
  dateRange,
  zoomLevel,
}: {
  position: number;
  dateRange: { start: Date; end: Date };
  zoomLevel: ZoomLevel;
}) => {
  const columnWidth = getColumnWidth(zoomLevel);
  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const timelineWidth = periods.length * columnWidth;
  const clampedPosition = Math.min(Math.max(position, 0), timelineWidth);

  if (zoomLevel === "weeks") {
    return addDays(dateRange.start, Math.round(clampedPosition / columnWidth));
  }

  const periodIndex = Math.min(
    periods.length - 1,
    Math.floor(clampedPosition / columnWidth),
  );
  const period = periods[Math.max(0, periodIndex)];
  const periodStart =
    zoomLevel === "months" ? startOfMonth(period) : startOfQuarter(period);
  const periodEnd =
    zoomLevel === "months" ? endOfMonth(period) : endOfQuarter(period);
  const daysInPeriod = differenceInDays(periodEnd, periodStart) + 1;
  const positionWithinPeriod = clampedPosition - periodIndex * columnWidth;
  const dayOffset = Math.min(
    daysInPeriod - 1,
    Math.floor((positionWithinPeriod / columnWidth) * daysInPeriod),
  );

  return addDays(periodStart, dayOffset);
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
