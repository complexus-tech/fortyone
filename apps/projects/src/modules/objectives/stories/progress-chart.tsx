"use client";
import { useMemo, useCallback } from "react";
import { format, isWeekend, parseISO } from "date-fns";
import {
  ComposedChart,
  Line,
  XAxis,
  Tooltip,
  ResponsiveContainer,
  ReferenceArea,
  Legend,
} from "recharts";
import { Box, Text } from "ui";
import { useTheme } from "next-themes";
import { useTerminology } from "@/hooks";

type ProgressChartData = {
  date: string;
  completed: number;
  inProgress: number;
  total: number;
};

type ProgressChartProps = {
  progressData: ProgressChartData[];
};

const CustomTooltip = ({
  active,
  payload,
  label,
}: {
  active?: boolean;
  payload?: {
    payload: {
      completed: number;
      inProgress: number;
      total: number;
      isWeekend: boolean;
    };
  }[];
  label?: string;
}) => {
  const { getTermDisplay } = useTerminology();
  if (active && payload?.length) {
    const data = payload[0].payload;
    return (
      <Box className="border-border bg-surface-elevated/60 text-foreground z-50 min-w-44 rounded-2xl border p-4 backdrop-blur-lg">
        <Text fontWeight="semibold">{label}</Text>
        <Box className="mb-0.1 mt-1 text-[#6366F1]">
          Completed: {data.completed}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        <Box className="text-warning mb-0.5">
          In Progress: {data.inProgress}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        <Box className="text-foreground mb-0.5">
          Total: {data.total}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        {data.isWeekend ? <Box className="opacity-60">Weekend</Box> : null}
      </Box>
    );
  }
  return null;
};

const CustomLegend = () => {
  const { resolvedTheme } = useTheme();

  return (
    <Box className="flex justify-center gap-6">
      <Box className="flex items-center gap-2">
        <Box className="h-1 w-3 bg-[#6366F1]" />
        <span className="text-foreground">Completed</span>
      </Box>
      <Box className="flex items-center gap-2">
        <Box className="bg-warning h-1 w-3" />
        <span className="text-foreground">In Progress</span>
      </Box>
      <Box className="flex items-center gap-2">
        <Box
          className="bg-foreground h-1 w-3"
          style={{
            borderColor: resolvedTheme === "dark" ? "#9CA3AF" : "#6B7280",
            opacity: 0.8,
          }}
        />
        <span className="text-foreground">Total</span>
      </Box>
    </Box>
  );
};

// Custom tick component for X-axis with dynamic text anchoring
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

  const oneThird = Math.floor(totalLength / 3);
  const twoThird = Math.floor((totalLength * 2) / 3);

  // Only show ticks at specific intervals (first, one-third, two-thirds, last)
  const shouldShow =
    index === 0 ||
    index === oneThird ||
    index === twoThird ||
    index === totalLength;

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

export const ProgressChart = ({ progressData }: ProgressChartProps) => {
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
        totalLength={progressData.length - 1}
      />
    ),
    [resolvedTheme, progressData.length],
  );

  const chartData = useMemo(() => {
    return progressData.map((item, index) => {
      const date = parseISO(item.date);
      const isWeekendDay = isWeekend(date);

      return {
        date: format(date, "MMM d"),
        day: index + 1,
        completed: item.completed,
        inProgress: item.inProgress,
        total: item.total,
        isWeekend: isWeekendDay,
        rawDate: item.date,
      };
    });
  }, [progressData]);

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

    // Handle case where period ends on weekend (last item)
    if (index === chartData.length - 1 && weekendStart) {
      weekendRanges.push({ start: weekendStart, end: item.date });
    }
  });

  return (
    <Box className="h-72 w-full">
      <ResponsiveContainer height="100%" width="100%">
        <ComposedChart
          data={chartData}
          margin={{
            top: 20,
            right: 0,
            left: 3,
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

          <Legend content={<CustomLegend />} />

          {/* Completed line */}
          <Line
            connectNulls={false}
            dataKey="completed"
            dot={false}
            stroke="#6366F1"
            strokeWidth={1.5}
            type="linear"
          />

          {/* In Progress line */}
          <Line
            connectNulls={false}
            dataKey="inProgress"
            dot={false}
            stroke="#eab308"
            strokeWidth={1.5}
            type="linear"
          />

          {/* Total line */}
          <Line
            connectNulls={false}
            dataKey="total"
            dot={false}
            stroke={resolvedTheme === "dark" ? "#9CA3AF" : "#6B7280"}
            strokeOpacity={0.8}
            strokeWidth={1.5}
            type="linear"
          />
        </ComposedChart>
      </ResponsiveContainer>
    </Box>
  );
};
