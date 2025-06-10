import React, { useCallback } from "react";
import { format, isWeekend } from "date-fns";
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
import { useTerminology } from "@/hooks";
import type { SprintAnalytics } from "../types";

type BurndownChartProps = {
  burndownData: SprintAnalytics["burndown"];
};

const CustomTooltip = ({
  active,
  payload,
  label,
}: {
  active?: boolean;
  payload?: {
    payload: { actual: number; ideal: number; isWeekend: boolean };
  }[];
  label?: string;
}) => {
  const { getTermDisplay } = useTerminology();
  if (active && payload?.length) {
    const data = payload[0].payload;
    return (
      <Box className="z-50 min-w-44 rounded-2xl border border-gray-100 bg-white/60 p-3 font-medium text-gray backdrop-blur dark:border-dark-50 dark:bg-dark-200/60 dark:text-gray-200">
        <Text fontWeight="semibold">{label}</Text>
        <Box className="mb-0.1 mt-1 text-warning">
          Remaining: {data.actual}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        <Box className="mb-0.5 text-[#6366F1]">
          Ideal: {data.ideal}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        {data.isWeekend ? <Box className="opacity-60">Weekend</Box> : null}
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

export const BurndownChart = ({ burndownData }: BurndownChartProps) => {
  const { resolvedTheme } = useTheme();

  // Memoized tick component to avoid recreation on every render
  const renderTick = useCallback(
    (props: {
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
    ),
    [resolvedTheme, burndownData.length],
  );

  // Transform the analytics data for the chart
  const chartData = burndownData.map((item, index) => {
    const date = new Date(item.date);
    const isWeekendDay = isWeekend(date);

    return {
      date: format(date, "MMM d"),
      day: index + 1,
      actual: item.remaining,
      ideal: item.ideal,
      isWeekend: isWeekendDay,
      rawDate: item.date,
    };
  });

  // Find weekend ranges for ReferenceArea
  const weekendRanges: { start: string; end: string }[] = [];
  let weekendStart: string | null = null;

  chartData.forEach((item, index) => {
    if (item.isWeekend && !weekendStart) {
      weekendStart = item.date;
    } else if (!item.isWeekend && weekendStart) {
      const prevItem = chartData[index - 1];
      weekendRanges.push({ start: weekendStart, end: prevItem.date });
      weekendStart = null;
    }

    // Handle case where sprint ends on weekend (last item)
    if (index === chartData.length - 1 && weekendStart) {
      weekendRanges.push({ start: weekendStart, end: item.date });
    }
  });

  return (
    <div className="h-60 w-full">
      <ResponsiveContainer height="100%" width="100%">
        <ComposedChart
          data={chartData}
          margin={{
            top: 20,
            right: 20,
            left: 2,
            bottom: 0,
          }}
        >
          <defs>
            <pattern
              height="6"
              id="weekendPattern"
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

          {weekendRanges.map((range, index) => (
            <ReferenceArea
              fill="url(#weekendPattern)"
              fillOpacity={0.6}
              key={index}
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
    </div>
  );
};
