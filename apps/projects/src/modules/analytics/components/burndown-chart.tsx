"use client";
import { Box, Flex, Text, Wrapper } from "ui";
import type { TooltipProps } from "recharts";
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  LineChart,
  Line,
} from "recharts";
import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { useStoryAnalytics } from "../hooks/story-analytics";
import type { BurndownPoint } from "../types";

type ChartDataItem = {
  date: string;
  remaining: number;
};

const CustomTooltip = ({
  active,
  payload,
  label,
}: TooltipProps<number, string>) => {
  if (!active || !payload?.length) {
    return null;
  }

  return (
    <Box className="z-50 min-w-28 rounded-[0.6rem] border border-gray-100 bg-white/80 px-3 py-3 text-[0.95rem] font-medium text-gray backdrop-blur dark:border-dark-50 dark:bg-dark-200 dark:text-gray-200">
      <Flex align="center" gap={2}>
        {label}
      </Flex>
      <Text className="mt-1 pl-0.5">{payload[0].value} remaining</Text>
    </Box>
  );
};

const formatDate = (date: string) => {
  const dateObj = new Date(date);
  return dateObj.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
};

export const BurndownChart = () => {
  const { resolvedTheme } = useTheme();
  const { data: storyAnalytics, isPending } = useStoryAnalytics();
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);

  useEffect(() => {
    if (storyAnalytics?.burndown.length) {
      const formattedData = storyAnalytics.burndown.map(
        (item: BurndownPoint) => ({
          date: formatDate(item.date),
          remaining: item.remaining,
        }),
      );
      setChartData(formattedData);
    }
  }, [storyAnalytics]);

  if (isPending) {
    return (
      <Wrapper>
        <Box className="mb-6">
          <Text className="mb-1" fontSize="lg">
            Burndown chart
          </Text>
          <Text color="muted">Story completion progress over time.</Text>
        </Box>
        <Box className="h-[220px] animate-pulse rounded bg-gray-200 dark:bg-dark-100" />
      </Wrapper>
    );
  }

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Burndown chart
        </Text>
        <Text color="muted">Story completion progress over time.</Text>
      </Box>

      <ResponsiveContainer height={220} width="100%">
        <LineChart
          data={chartData}
          margin={{ top: 20, right: 10, left: -35, bottom: 0 }}
        >
          <CartesianGrid
            stroke={resolvedTheme === "dark" ? "#222" : "#E0E0E0"}
            strokeDasharray="3 3"
            vertical={false}
          />
          <XAxis
            axisLine={{
              stroke: resolvedTheme === "dark" ? "#222" : "#E0E0E0",
            }}
            dataKey="date"
            tick={{ fontSize: 12 }}
          />
          <YAxis axisLine={false} tick={{ fontSize: 12 }} tickLine={false} />
          <Tooltip
            content={<CustomTooltip />}
            cursor={{
              stroke: resolvedTheme === "dark" ? "#444" : "#ccc",
              strokeWidth: 1,
            }}
          />
          <Line
            activeDot={{
              r: 4,
              fill: "#6366F1",
              strokeWidth: 2,
              stroke:
                resolvedTheme === "dark"
                  ? "rgba(255, 255, 255, 0.8)"
                  : "#6366F1",
            }}
            dataKey="remaining"
            dot={{
              r: 3,
              fill: "#6366F1",
              strokeWidth: 2,
              stroke:
                resolvedTheme === "dark"
                  ? "rgba(255, 255, 255, 0.8)"
                  : "#6366F1",
            }}
            stroke="#6366F1"
            strokeWidth={2}
            type="monotone"
          />
        </LineChart>
      </ResponsiveContainer>
    </Wrapper>
  );
};
