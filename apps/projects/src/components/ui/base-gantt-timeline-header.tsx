"use client";

import {
  endOfMonth,
  endOfQuarter,
  format,
  getWeek,
  isSameWeek,
  isWeekend,
  startOfQuarter,
} from "date-fns";
import { cn } from "lib";
import { Box, Flex, Text } from "ui";
import type { GanttDateRange } from "./base-gantt-types";
import {
  getColumnWidth,
  getTimePeriodsForZoom,
  type ZoomLevel,
} from "./base-gantt-utils";

type WeekSpan = {
  week: string;
  month: string;
  startIndex: number;
  span: number;
};

const getWeekSpans = (days: Date[]): WeekSpan[] => {
  if (days.length === 0) return [];

  const spans: WeekSpan[] = [];
  let startIndex = 0;

  for (let index = 0; index < days.length; index++) {
    const currentDay = days[index];
    const nextDay = days[index + 1];
    const isEndOfWeek =
      index === days.length - 1 ||
      !isSameWeek(currentDay, nextDay, { weekStartsOn: 0 });

    if (!isEndOfWeek) continue;

    const weekStart = days[startIndex];
    spans.push({
      month: format(weekStart, "MMM yyyy"),
      span: index - startIndex + 1,
      startIndex,
      week: `Week ${getWeek(weekStart, { weekStartsOn: 0 })}`,
    });
    startIndex = index + 1;
  }

  return spans;
};

const PeriodPrimaryHeader = ({
  left,
  right,
}: {
  left: string;
  right: string;
}) => (
  <Flex align="center" className="h-5 min-h-0" justify="between">
    <Text className="text-[0.9rem]" color="muted" fontWeight="semibold">
      {left}
    </Text>
    <Text
      className="text-[0.9rem] opacity-60"
      color="muted"
      fontWeight="semibold"
    >
      {right}
    </Text>
  </Flex>
);

const WeekTimelineHeader = ({
  periods,
  columnWidth,
}: {
  periods: Date[];
  columnWidth: number;
}) => (
  <>
    <Box className="border-border dark:border-border/45 border-b-[0.5px]">
      <Flex>
        {getWeekSpans(periods).map(({ week, month, span, startIndex }) => (
          <Box
            className="border-border dark:border-border/45 border-r-[0.5px] px-2 py-1.5 text-left"
            key={`${month}-${week}-${startIndex}`}
            style={{ width: `${(span / periods.length) * 100}%` }}
          >
            <PeriodPrimaryHeader left={month} right={week} />
          </Box>
        ))}
      </Flex>
    </Box>
    <Flex>
      {periods.map((day) => (
        <Box
          className={cn(
            "border-border dark:border-border/45 h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] px-1 py-1 text-center",
            { "bg-surface-muted": isWeekend(day) },
          )}
          key={day.getTime()}
          style={{ minWidth: `${columnWidth}px` }}
        >
          <Flex align="center" className="px-1" justify="between">
            <Text color="muted" fontSize="sm">
              {format(day, "d")}
            </Text>
            <Text color="muted" fontSize="sm">
              {format(day, "eeeee")}
            </Text>
          </Flex>
        </Box>
      ))}
    </Flex>
  </>
);

const PeriodTimelineHeader = ({
  periods,
  columnWidth,
  zoomLevel,
}: {
  periods: Date[];
  columnWidth: number;
  zoomLevel: Exclude<ZoomLevel, "weeks">;
}) => (
  <>
    <Box className="border-border dark:border-border/45 border-b-[0.5px]">
      <Flex>
        {periods.map((period) => (
          <Box
            className="border-border dark:border-border/45 border-r-[0.5px] px-2 py-1.5 text-left"
            key={period.getTime()}
            style={{ minWidth: `${columnWidth}px` }}
          >
            <PeriodPrimaryHeader
              left={
                zoomLevel === "months"
                  ? format(period, "MMM")
                  : `Q${Math.ceil((period.getMonth() + 1) / 3)}`
              }
              right={format(period, "yyyy")}
            />
          </Box>
        ))}
      </Flex>
    </Box>
    <Flex>
      {periods.map((period) => {
        const range =
          zoomLevel === "months"
            ? { start: period, end: endOfMonth(period) }
            : { start: startOfQuarter(period), end: endOfQuarter(period) };

        return (
          <Box
            className="border-border dark:border-border/45 h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] px-1 py-1 text-center"
            key={period.getTime()}
            style={{ minWidth: `${columnWidth}px` }}
          >
            <Flex align="center" className="px-1" justify="between">
              <Text color="muted" fontSize="sm">
                {format(range.start, zoomLevel === "months" ? "d" : "MMM")}
              </Text>
              <Text color="muted" fontSize="sm">
                {format(range.end, zoomLevel === "months" ? "d" : "MMM")}
              </Text>
            </Flex>
          </Box>
        );
      })}
    </Flex>
  </>
);

export const BaseGanttTimelineHeader = ({
  dateRange,
  zoomLevel,
}: {
  dateRange: GanttDateRange;
  zoomLevel: ZoomLevel;
}) => {
  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const columnWidth = getColumnWidth(zoomLevel);

  return (
    <Box
      className="border-border bg-background dark:border-border/45 sticky top-0 z-10 h-16 border-b-[0.5px]"
      style={{ minWidth: `${periods.length * columnWidth}px` }}
    >
      <Box className="h-8 w-full">
        {zoomLevel === "weeks" ? (
          <WeekTimelineHeader columnWidth={columnWidth} periods={periods} />
        ) : (
          <PeriodTimelineHeader
            columnWidth={columnWidth}
            periods={periods}
            zoomLevel={zoomLevel}
          />
        )}
      </Box>
    </Box>
  );
};
