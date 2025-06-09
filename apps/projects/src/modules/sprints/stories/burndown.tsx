import React from "react";
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
      <Box className="z-50 min-w-44 rounded-2xl border border-gray-100 bg-white/80 p-3 font-medium text-gray backdrop-blur dark:border-dark-50 dark:bg-dark-200/60 dark:text-gray-200">
        <Text fontWeight="semibold">{label}</Text>
        <Box className="mb-0.1 mt-1 text-warning">
          Actual: {data.actual}{" "}
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

export const BurndownChart = ({ burndownData }: BurndownChartProps) => {
  const { resolvedTheme } = useTheme();

  // Transform the analytics data for the chart
  const chartData = burndownData.map((item, index) => {
    const date = new Date(item.date);
    const isWeekendDay = isWeekend(date);

    // For weekends, use the ideal value from the last working day
    let idealValue = item.ideal;
    if (isWeekendDay && index > 0) {
      // Find the last working day's ideal value
      for (let i = index - 1; i >= 0; i--) {
        const prevDate = new Date(burndownData[i].date);
        if (!isWeekend(prevDate)) {
          idealValue = burndownData[i].ideal;
          break;
        }
      }
    }

    return {
      date: format(date, "MMM d"),
      day: index + 1,
      actual: item.remaining,
      ideal: idealValue,
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
            left: -35,
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
            tick={{ fontSize: 12 }}
            tickFormatter={(value: string, index) => {
              const totalLength = chartData.length - 1;
              const quarter = Math.floor(totalLength / 4);
              const middle = Math.floor(totalLength / 2);
              const threeQuarter = Math.floor((totalLength * 3) / 4);

              if (
                index === 0 ||
                index === quarter ||
                index === middle ||
                index === threeQuarter ||
                index === totalLength
              ) {
                return value;
              }
              return "";
            }}
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
            type="linear"
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
