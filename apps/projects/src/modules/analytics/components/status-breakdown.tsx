"use client";
import { Flex, Text, Wrapper, Box } from "ui";
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from "recharts";
import { useMemo } from "react";
import type { TooltipProps } from "recharts";
import { useTerminology } from "@/hooks";
import { useStoryAnalytics } from "../hooks/story-analytics";
import type { StatusBreakdownItem } from "../types";
import { useAppliedFilters } from "../hooks/filters";
import { StatusBreakdownSkeleton } from "./status-breakdown-skeleton";

// Define colors for different statuses
const COLORS = [
  "#6366F1",
  "#EA6060",
  "#002F61",
  "#22c55e",
  "#eab308",
  "#06b6d4",
  "#f43f5e",
];

const CustomTooltip = ({ active, payload }: TooltipProps<number, string>) => {
  const { getTermDisplay } = useTerminology();
  if (!active || !payload?.length) {
    return null;
  }

  return (
    <Box className="border-border bg-surface-elevated/80 text-foreground relative z-50 min-w-28 rounded-lg border px-3 py-3 text-[0.95rem] font-medium backdrop-blur">
      <Flex align="center" gap={2}>
        {payload[0].name}
      </Flex>
      <Text className="mt-1 pl-0.5">
        {payload[0].value}{" "}
        {getTermDisplay("storyTerm", {
          variant: payload[0].value === 1 ? "singular" : "plural",
        })}
      </Text>
    </Box>
  );
};

export const StatusBreakdown = () => {
  const { getTermDisplay } = useTerminology();
  const filters = useAppliedFilters();
  const { data: storyAnalytics, isPending } = useStoryAnalytics(filters);
  const chartData = storyAnalytics?.statusBreakdown ?? [];
  const totalCount = useMemo(
    () =>
      chartData.reduce(
        (sum: number, status: StatusBreakdownItem) => sum + status.count,
        0,
      ),
    [chartData],
  );

  if (isPending) {
    return <StatusBreakdownSkeleton />;
  }

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text fontSize="lg">Status breakdown</Text>
        <Text color="muted">Stories by status across teams.</Text>
      </Box>

      <Box>
        <Box className="relative isolate">
          <Box className="absolute top-1/2 left-1/2 z-0 -translate-x-1/2 -translate-y-1/2 transform text-center">
            <Text fontSize="3xl">{totalCount}</Text>
            <Text color="muted">
              Total{" "}
              {getTermDisplay("storyTerm", {
                variant: "plural",
              })}
            </Text>
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
                nameKey="statusName"
                outerRadius={80}
                paddingAngle={2}
                stroke="none"
              >
                {chartData.map((entry, index) => (
                  <Cell
                    fill={COLORS[index % COLORS.length]}
                    key={`cell-${entry.statusName}`}
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
            <Flex align="center" gap={1} key={entry.statusName}>
              <Box
                className="size-4 rounded"
                style={{ backgroundColor: COLORS[index % COLORS.length] }}
              />
              <Text>{entry.statusName}</Text>
            </Flex>
          ))}
        </Flex>
      </Box>
    </Wrapper>
  );
};
