"use client";
import { Flex, Text, Wrapper, Box } from "ui";
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from "recharts";
import { useMemo } from "react";
import type { TooltipProps } from "recharts";
import { useObjectiveProgress } from "../hooks/objective-progress";
import type { HealthDistributionItem } from "../types";
import { useAppliedFilters } from "../hooks/filters";
import { ObjectiveHealthSkeleton } from "./objective-health-skeleton";

// Define colors for different health statuses
const HEALTH_COLORS = {
  "On Track": "#22c55e",
  "At Risk": "#eab308",
  "Off Track": "#EA6060",
  "Not Started": "#6b7280",
  Completed: "#10b981",
  Blocked: "#f97316",
  Planning: "#6366F1",
  Review: "#8b5cf6",
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
    <Box className="border-border bg-surface-elevated/80 text-foreground relative z-50 min-w-28 rounded-lg border px-3 py-3 text-[0.95rem] font-medium backdrop-blur">
      <Flex align="center" gap={2}>
        {payload[0].name}
      </Flex>
      <Text className="mt-1 pl-0.5">{payload[0].value} objectives</Text>
    </Box>
  );
};

export const ObjectiveHealth = () => {
  const filters = useAppliedFilters();
  const { data: objectiveProgress, isPending } = useObjectiveProgress(filters);
  const chartData = objectiveProgress?.healthDistribution ?? [];
  const totalCount = useMemo(
    () =>
      chartData.reduce(
        (sum: number, health: HealthDistributionItem) => sum + health.count,
        0,
      ),
    [chartData],
  );

  if (isPending) {
    return <ObjectiveHealthSkeleton />;
  }

  const getColor = (status: string, index: number) => {
    return (
      HEALTH_COLORS[status as keyof typeof HEALTH_COLORS] ||
      FALLBACK_COLORS[index % FALLBACK_COLORS.length]
    );
  };

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text fontSize="lg">Objective health</Text>
        <Text color="muted">Health status distribution of objectives.</Text>
      </Box>

      <Box>
        <Box className="relative isolate">
          <Box className="absolute top-1/2 left-1/2 z-0 -translate-x-1/2 -translate-y-1/2 transform text-center">
            <Text fontSize="3xl">{totalCount}</Text>
            <Text color="muted">Total objectives</Text>
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
                    key={`cell-${entry.status}`}
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
            <Flex align="center" gap={1} key={entry.status}>
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
