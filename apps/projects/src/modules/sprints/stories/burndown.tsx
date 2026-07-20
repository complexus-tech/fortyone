import React from "react";
import { format } from "date-fns";
import {
  Line,
  XAxis,
  ResponsiveContainer,
  ComposedChart,
  Tooltip,
  ReferenceArea,
} from "recharts";
import { useTheme } from "next-themes";
import { Box, Text } from "ui";
import { cn } from "lib";
import { useTerminology } from "@/hooks";
import type { SprintAnalytics } from "../types";

const DEFAULT_WORKING_DAYS = [1, 2, 3, 4, 5];

type BurndownChartProps = {
  burndownData: SprintAnalytics["burndown"];
  workingDays?: SprintAnalytics["workingDays"];
  className?: string;
};

const CustomTooltip = ({
  active,
  payload,
  label,
}: {
  active?: boolean;
  payload?: {
    payload: { actual: number; ideal: number; isNonWorkingDay: boolean };
  }[];
  label?: string;
}) => {
  const { getTermDisplay } = useTerminology();
  if (active && payload?.length) {
    const data = payload[0].payload;
    return (
      <Box className="border-border/60 bg-surface-elevated/60 text-foreground z-50 min-w-44 rounded-2xl border-[0.5px] p-4 backdrop-blur-lg">
        <Text fontWeight="semibold">{label}</Text>
        <Box className="mb-0.1 text-warning mt-1">
          Remaining: {data.actual}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        <Box className="mb-0.5 text-[#6366F1]">
          Ideal: {data.ideal}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        {data.isNonWorkingDay ? (
          <Box className="opacity-60">Non-working day</Box>
        ) : null}
      </Box>
    );
  }
  return null;
};

// Custom tick component defined outside the main component
const CustomXAxisTick = ({
  x,
  y,
  payload,
  index,
  totalLength,
  isDark,
}: {
  x: number;
  y: number;
  payload: { value: string };
  index: number;
  totalLength: number;
  isDark: boolean;
}) => {
  // Determine text anchor based on position
  let textAnchor: "start" | "middle" | "end" = "middle";
  if (index === 0) {
    textAnchor = "start";
  } else if (index === totalLength) {
    textAnchor = "end";
  }

  const middle = Math.floor(totalLength / 2);

  // Only show ticks at specific intervals (first, middle, last)
  const shouldShow = index === 0 || index === middle || index === totalLength;

  if (!shouldShow) {
    return null;
  }

  return (
    <text
      dy={16}
      fill={isDark ? "#9CA3AF" : "#6B7280"}
      fontSize={12}
      textAnchor={textAnchor}
      x={x}
      y={y}
    >
      {payload.value}
    </text>
  );
};

export const BurndownChart = ({
  burndownData,
  workingDays = DEFAULT_WORKING_DAYS,
  className,
}: BurndownChartProps) => {
  const { resolvedTheme } = useTheme();

  const renderTick = (props: {
    x: number;
    y: number;
    payload: { value: string };
    index: number;
  }) => (
    <CustomXAxisTick
      {...props}
      isDark={resolvedTheme === "dark"}
      totalLength={burndownData.length - 1}
    />
  );

  // Transform the analytics data for the chart
  const chartData = burndownData.map((item, index) => {
    const date = new Date(item.date);
    const utcWeekday = date.getUTCDay();
    const isoWeekday = utcWeekday === 0 ? 7 : utcWeekday;
    const isNonWorkingDay = !workingDays.includes(isoWeekday);

    return {
      date: format(date, "MMM d"),
      day: index + 1,
      actual: item.remaining,
      ideal: item.ideal,
      isNonWorkingDay,
      rawDate: item.date,
    };
  });

  // Find non-working ranges for ReferenceArea
  const nonWorkingRanges: { start: string; end: string }[] = [];
  let nonWorkingStart: string | null = null;

  chartData.forEach((item, index) => {
    if (item.isNonWorkingDay && !nonWorkingStart) {
      nonWorkingStart = item.date;
    } else if (!item.isNonWorkingDay && nonWorkingStart) {
      const prevItem = chartData[index - 1];
      nonWorkingRanges.push({
        start: nonWorkingStart,
        end: prevItem.date,
      });
      nonWorkingStart = null;
    }

    if (index === chartData.length - 1 && nonWorkingStart) {
      nonWorkingRanges.push({ start: nonWorkingStart, end: item.date });
    }
  });

  return (
    <Box className={cn("h-64 w-full", className)}>
      <ResponsiveContainer height="100%" width="100%">
        <ComposedChart
          data={chartData}
          margin={{
            top: 20,
            right: 0,
            left: 2,
            bottom: 0,
          }}
        >
          <defs>
            <pattern
              height="6"
              id="nonWorkingPattern"
              patternTransform="rotate(45)"
              patternUnits="userSpaceOnUse"
              width="6"
            >
              <rect
                fill={resolvedTheme === "dark" ? "#6b7280" : "#6b7280"}
                fillOpacity="0.4"
                height="6"
                width="1"
              />
              <rect fill="transparent" height="6" width="5" x="1" />
            </pattern>
          </defs>

          {nonWorkingRanges.map((range) => (
            <ReferenceArea
              fill="url(#nonWorkingPattern)"
              fillOpacity={0.6}
              key={`${range.start}-${range.end}`}
              x1={range.start}
              x2={range.end}
            />
          ))}
          <XAxis
            axisLine={{
              stroke: resolvedTheme === "dark" ? "#222" : "#E0E0E0",
            }}
            dataKey="date"
            interval={0}
            tick={renderTick}
            tickLine={{
              stroke: resolvedTheme === "dark" ? "#333" : "#E0E0E0",
            }}
          />

          <Tooltip
            content={<CustomTooltip />}
            cursor={{
              fill:
                resolvedTheme === "dark"
                  ? "rgba(255, 255, 255, 0.03)"
                  : "rgba(0, 0, 0, 0.05)",
            }}
          />

          <Line
            connectNulls={false}
            dataKey="ideal"
            dot={false}
            stroke="#6366F1"
            strokeDasharray="8 4"
            strokeOpacity={0.6}
            strokeWidth={2}
            type="monotone"
          />
          <Line
            connectNulls={false}
            dataKey="actual"
            dot={false}
            stroke="#eab308"
            strokeWidth={2}
            type="monotone"
          />
        </ComposedChart>
      </ResponsiveContainer>
    </Box>
  );
};
