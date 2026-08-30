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
  subDays,
} from "date-fns";

export type ZoomLevel = "weeks" | "months" | "quarters";
export type OffscreenDirection = "left" | "right";
export type GanttVirtualRow = {
  index: number;
  size: number;
  start: number;
};

export const getGanttVirtualRows = ({
  itemCount,
  rowHeight,
  scrollTop,
  viewportHeight,
  headerHeight = 64,
  overscan = 6,
  pinnedIndices = [],
}: {
  itemCount: number;
  rowHeight: number;
  scrollTop: number;
  viewportHeight: number;
  headerHeight?: number;
  overscan?: number;
  pinnedIndices?: readonly number[];
}) => {
  const totalSize = itemCount * rowHeight;
  if (itemCount === 0) return { rows: [], totalSize };

  const visibleStart = Math.max(0, scrollTop - headerHeight);
  const visibleEnd = Math.max(
    visibleStart + rowHeight,
    scrollTop + viewportHeight - headerHeight,
  );
  const firstVisibleIndex = Math.min(
    itemCount - 1,
    Math.floor(visibleStart / rowHeight),
  );
  const lastVisibleIndex = Math.min(
    itemCount - 1,
    Math.max(firstVisibleIndex, Math.ceil(visibleEnd / rowHeight) - 1),
  );
  const startIndex = Math.max(0, firstVisibleIndex - overscan);
  const endIndex = Math.min(itemCount, lastVisibleIndex + overscan + 1);
  const renderedIndices = new Set<number>();

  for (let index = startIndex; index < endIndex; index++) {
    renderedIndices.add(index);
  }
  for (const index of pinnedIndices) {
    if (index >= 0 && index < itemCount) renderedIndices.add(index);
  }

  return {
    rows: Array.from(renderedIndices)
      .sort((left, right) => left - right)
      .map((index) => ({
        index,
        size: rowHeight,
        start: index * rowHeight,
      })),
    totalSize,
  };
};

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

export const getGanttDateRange = <
  T extends { startDate?: string | null; endDate?: string | null },
>(
  centerDate: Date,
  items: T[],
  zoomLevel: ZoomLevel,
) => {
  const viewportDays = zoomLevel === "weeks" ? 365 : 1460;
  const paddingDays = zoomLevel === "weeks" ? 30 : 120;
  const halfViewport = Math.floor(viewportDays / 2);
  const start = subDays(centerDate, halfViewport);
  const end = addDays(centerDate, halfViewport);

  start.setHours(0, 0, 0, 0);
  end.setHours(0, 0, 0, 0);

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
    { start, end },
  );
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
