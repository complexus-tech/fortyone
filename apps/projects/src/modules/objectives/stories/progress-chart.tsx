"use client";
import { useMemo } from "react";
import { format, isWeekend, parseISO } from "date-fns";
import {
  ComposedChart,
  Line,
  XAxis,
  YAxis,
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
      <Box className="z-50 min-w-44 rounded-2xl border border-gray-100 bg-white/60 p-3 font-medium text-gray backdrop-blur dark:border-dark-50 dark:bg-dark-200/60 dark:text-gray-200">
        <Text fontWeight="semibold">{label}</Text>
        <Box className="mb-0.1 mt-1 text-[#6366F1]">
          Completed: {data.completed}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        <Box className="mb-0.1 text-warning">
          In Progress: {data.inProgress}{" "}
          {getTermDisplay("storyTerm", { variant: "plural" })}
        </Box>
        <Box className="text-gray-600 mb-0.5 dark:text-gray-300">
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
        <Box className="h-0.5 w-3 bg-[#6366F1]" />
        <span className="text-gray dark:text-gray-300">Completed</span>
      </Box>
      <Box className="flex items-center gap-2">
        <Box className="h-0.5 w-3 bg-[#eab308]" />
        <span className="text-gray dark:text-gray-300">In Progress</span>
      </Box>
      <Box className="flex items-center gap-2">
        <Box
          className="h-0.5 w-3 border-b-2 border-dashed"
          style={{
            borderColor: resolvedTheme === "dark" ? "#9CA3AF" : "#6B7280",
            opacity: 0.6,
          }}
        />
        <span className="text-gray dark:text-gray-300">Total</span>
      </Box>
    </Box>
  );
};

export const ProgressChart = ({ progressData }: ProgressChartProps) => {
  const { resolvedTheme } = useTheme();

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

  const maxValue = Math.max(
    ...chartData.map((d) => Math.max(d.completed, d.inProgress, d.total)),
  );
  const yAxisMax = Math.ceil(maxValue * 1.1);

  return (
    <div className="h-64 w-full">
      <ResponsiveContainer height="100%" width="100%">
        <ComposedChart
          data={chartData}
          margin={{
            top: 20,
            right: 20,
            left: -40,
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
              const oneThird = Math.floor(totalLength / 3);
              const twoThird = Math.floor((totalLength * 2) / 3);

              if (
                index === 0 ||
                index === oneThird ||
                index === twoThird ||
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

          <YAxis
            axisLine={{
              stroke: resolvedTheme === "dark" ? "#222" : "#E0E0E0",
            }}
            domain={[0, yAxisMax]}
            tick={{
              fontSize: 12,
              fill: resolvedTheme === "dark" ? "#9CA3AF" : "#6B7280",
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

          <Legend content={<CustomLegend />} />

          {/* Completed line */}
          <Line
            connectNulls={false}
            dataKey="completed"
            dot={false}
            stroke="#6366F1"
            strokeWidth={2}
            type="linear"
          />

          {/* In Progress line */}
          <Line
            connectNulls={false}
            dataKey="inProgress"
            dot={false}
            stroke="#eab308"
            strokeOpacity={0.8}
            strokeWidth={2}
            type="linear"
          />

          {/* Total line */}
          <Line
            connectNulls={false}
            dataKey="total"
            dot={false}
            stroke={resolvedTheme === "dark" ? "#9CA3AF" : "#6B7280"}
            strokeDasharray="8 4"
            strokeOpacity={0.6}
            strokeWidth={2}
            type="linear"
          />
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
};
