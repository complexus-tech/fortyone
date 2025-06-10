"use client";
import { Flex, Text, Wrapper, Box } from "ui";
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from "recharts";
import { useState, useEffect } from "react";
import type { TooltipProps } from "recharts";
import { useSprintAnalytics } from "../hooks/sprint-analytics";
import type { SprintHealthItem } from "../types";
import { SprintHealthSkeleton } from "./sprint-health-skeleton";

// Define colors for different sprint statuses
const SPRINT_COLORS = {
  Active: "#22c55e",
  Planning: "#6366F1",
  Complete: "#10b981",
  "On Hold": "#eab308",
  Cancelled: "#EA6060",
  Draft: "#6b7280",
  Review: "#8b5cf6",
  Archived: "#374151",
};

// Fallback color array for unknown statuses
const FALLBACK_COLORS = [
  "#6366F1",
  "#22c55e",
  "#eab308",
  "#EA6060",
  "#06b6d4",
  "#f43f5e",
  "#8b5cf6",
  "#f97316",
];

const CustomTooltip = ({ active, payload }: TooltipProps<number, string>) => {
  if (!active || !payload?.length) {
    return null;
  }

  return (
    <Box className="relative z-50 min-w-28 rounded-[0.6rem] border border-gray-100 bg-white/80 px-3 py-3 text-[0.95rem] font-medium text-gray backdrop-blur dark:border-dark-50 dark:bg-dark-200 dark:text-gray-200">
      <Flex align="center" gap={2}>
        {payload[0].name}
      </Flex>
      <Text className="mt-1 pl-0.5">{payload[0].value} sprints</Text>
    </Box>
  );
};

export const SprintHealth = () => {
  const { data: sprintAnalytics, isPending } = useSprintAnalytics();
  const [chartData, setChartData] = useState<SprintHealthItem[]>([]);
  const [totalCount, setTotalCount] = useState(0);

  useEffect(() => {
    if (sprintAnalytics?.sprintHealth.length) {
      setChartData(sprintAnalytics.sprintHealth);
      const total = sprintAnalytics.sprintHealth.reduce(
        (sum: number, health: SprintHealthItem) => sum + health.count,
        0,
      );
      setTotalCount(total);
    }
  }, [sprintAnalytics]);

  if (isPending) {
    return <SprintHealthSkeleton />;
  }

  const getColor = (status: string, index: number) => {
    return (
      SPRINT_COLORS[status as keyof typeof SPRINT_COLORS] ||
      FALLBACK_COLORS[index % FALLBACK_COLORS.length]
    );
  };

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text fontSize="lg">Sprint health</Text>
        <Text color="muted">Status distribution of active sprints.</Text>
      </Box>

      <Box>
        <Box className="relative isolate">
          <Box className="absolute left-1/2 top-1/2 z-0 -translate-x-1/2 -translate-y-1/2 transform text-center">
            <Text fontSize="3xl">{totalCount}</Text>
            <Text color="muted">Total sprints</Text>
          </Box>
          <ResponsiveContainer height={160} width="100%">
            <PieChart className="relative">
              <Pie
                cornerRadius={4}
                cx="50%"
                cy="50%"
                data={chartData}
                dataKey="count"
                fill="#8884d8"
                innerRadius={60}
                labelLine={false}
                nameKey="status"
                outerRadius={80}
                paddingAngle={2}
                stroke="none"
              >
                {chartData.map((entry, index) => (
                  <Cell
                    fill={getColor(entry.status, index)}
                    key={`cell-${index}`}
                    stroke="none"
                  />
                ))}
              </Pie>
              <Tooltip content={<CustomTooltip />} />
            </PieChart>
          </ResponsiveContainer>
        </Box>
        <Flex className="line-clamp-2 h-[60px] pt-3" gap={3} wrap>
          {chartData.map((entry, index) => (
            <Flex align="center" gap={1} key={`${entry.status}-${index}`}>
              <Box
                className="size-4 rounded"
                style={{ backgroundColor: getColor(entry.status, index) }}
              />
              <Text>{entry.status}</Text>
            </Flex>
          ))}
        </Flex>
      </Box>
    </Wrapper>
  );
};
