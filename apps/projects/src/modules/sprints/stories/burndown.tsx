import { useId } from "react";
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

type BurndownChartRow = {
  actual: number;
  date: string;
  ideal: number;
  isNonWorkingDay: boolean;
  label: string;
};

const CustomTooltip = ({
  active,
  payload,
}: {
  active?: boolean;
  payload?: {
    payload: BurndownChartRow;
  }[];
}) => {
  const { getTermDisplay } = useTerminology();
  if (active && payload?.length) {
    const data = payload[0].payload;
    return (
      <Box className="border-border/40 bg-surface-elevated/85 text-foreground z-50 min-w-44 rounded-2xl border-[0.5px] p-4 shadow-lg shadow-black/5 backdrop-blur-xl dark:border-white/[0.08] dark:shadow-black/20">
        <Text className="text-base font-semibold">{data.label}</Text>
        <Box className="text-warning mt-2 text-[0.95rem] font-medium">
          Remaining: {data.actual}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        <Box className="mt-1 text-[0.95rem] font-medium text-[#6366F1]">
          Ideal: {data.ideal}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        {data.isNonWorkingDay ? (
          <Box className="mt-2 text-sm opacity-55">Non-working day</Box>
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
      fontSize={13}
      textAnchor={textAnchor}
      x={x}
      y={y}
    >
      {format(new Date(payload.value), "MMM d")}
    </text>
  );
};

export const BurndownChart = ({
  burndownData,
  workingDays = DEFAULT_WORKING_DAYS,
  className,
}: BurndownChartProps) => {
  const { resolvedTheme } = useTheme();
  const patternId = `non-working-${useId().replaceAll(":", "")}`;
  const isDark = resolvedTheme === "dark";

  const renderTick = (props: {
    x: number;
    y: number;
    payload: { value: string };
    index: number;
  }) => (
    <CustomXAxisTick
      {...props}
      isDark={isDark}
      totalLength={burndownData.length - 1}
    />
  );

  // Transform the analytics data for the chart
  const chartData = burndownData.map((item) => {
    const date = new Date(item.date);
    const utcWeekday = date.getUTCDay();
    const isoWeekday = utcWeekday === 0 ? 7 : utcWeekday;
    const isNonWorkingDay = !workingDays.includes(isoWeekday);

    return {
      date: item.date,
      actual: item.remaining,
      ideal: item.ideal,
      isNonWorkingDay,
      label: format(date, "MMM d"),
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
              id={patternId}
              patternTransform="rotate(45)"
              patternUnits="userSpaceOnUse"
              width="6"
            >
              <rect
                fill={isDark ? "#71717A" : "#94A3B8"}
                fillOpacity={isDark ? 0.25 : 0.22}
                height="6"
                width="1"
              />
              <rect fill="transparent" height="6" width="5" x="1" />
            </pattern>
          </defs>

          {nonWorkingRanges.map((range) => (
            <ReferenceArea
              fill={`url(#${patternId})`}
              fillOpacity={0.7}
              key={`${range.start}-${range.end}`}
              x1={range.start}
              x2={range.end}
            />
          ))}
          <XAxis
            axisLine={{
              stroke: isDark
                ? "rgba(255, 255, 255, 0.09)"
                : "rgba(15, 23, 42, 0.12)",
            }}
            dataKey="date"
            interval={0}
            tick={renderTick}
            tickLine={{
              stroke: isDark
                ? "rgba(255, 255, 255, 0.12)"
                : "rgba(15, 23, 42, 0.12)",
            }}
          />

          <Tooltip
            allowEscapeViewBox={{ x: false, y: true }}
            content={<CustomTooltip />}
            cursor={{
              stroke: isDark
                ? "rgba(255, 255, 255, 0.16)"
                : "rgba(15, 23, 42, 0.18)",
              strokeDasharray: "3 4",
              strokeWidth: 1,
            }}
            wrapperStyle={{ outline: "none" }}
          />

          <Line
            activeDot={{
              fill: "#6366F1",
              r: 6,
              stroke: isDark ? "#18181B" : "#FFFFFF",
              strokeWidth: 3,
            }}
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
            activeDot={{
              fill: "#EAB308",
              r: 6,
              stroke: isDark ? "#18181B" : "#FFFFFF",
              strokeWidth: 3,
            }}
            connectNulls={false}
            dataKey="actual"
            dot={false}
            stroke="#EAB308"
            strokeWidth={2}
            type="monotone"
          />
        </ComposedChart>
      </ResponsiveContainer>
    </Box>
  );
};
