"use client";
import { Box, Text, Wrapper } from "ui";
import type { TooltipProps } from "recharts";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from "recharts";
import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { useSprintAnalytics } from "../hooks/sprint-analytics";
import type { TeamAllocationItem } from "../types";
import { useAppliedFilters } from "../hooks/filters";
import { TeamAllocationSkeleton } from "./team-allocation-skeleton";

// Define colors for different teams
const TEAM_COLORS = [
  "#6366F1",
  "#22c55e",
  "#eab308",
  "#EA6060",
  "#06b6d4",
  "#f43f5e",
  "#8b5cf6",
  "#f97316",
];

type ChartDataItem = {
  teamName: string;
  activeSprints: number;
  totalStories: number;
  completedStories: number;
};

const CustomTooltip = ({
  active,
  payload,
  label,
}: TooltipProps<number, string>) => {
  if (!active || !payload?.length) {
    return null;
  }

  const data = payload[0].payload;

  return (
    <Box className="z-50 min-w-32 rounded-[0.6rem] border border-border bg-surface-elevated/80 px-3 py-3 text-[0.95rem] font-medium text-foreground backdrop-blur">
      <Text className="mb-2 font-medium">{label}</Text>
      <Box className="space-y-1">
        <Text>Active sprints: {data.activeSprints}</Text>
        <Text>Total stories: {data.totalStories}</Text>
        <Text>Completed: {data.completedStories}</Text>
      </Box>
    </Box>
  );
};

export const TeamAllocation = () => {
  const { resolvedTheme } = useTheme();
  const filters = useAppliedFilters();
  const { data: sprintAnalytics, isPending } = useSprintAnalytics(filters);
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);

  useEffect(() => {
    if (sprintAnalytics?.teamAllocation.length) {
      const formattedData = sprintAnalytics.teamAllocation.map(
        (item: TeamAllocationItem) => ({
          teamName: item.teamName,
          activeSprints: item.activeSprints,
          totalStories: item.totalStories,
          completedStories: item.completedStories,
        }),
      );
      setChartData(formattedData);
    }
  }, [sprintAnalytics]);

  if (isPending) {
    return <TeamAllocationSkeleton />;
  }

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Team allocation
        </Text>
        <Text color="muted">Sprint allocation and progress by team.</Text>
      </Box>

      <ResponsiveContainer height={220} width="100%">
        <BarChart
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
            dataKey="teamName"
            tick={{ fontSize: 12 }}
          />
          <YAxis axisLine={false} tick={{ fontSize: 12 }} tickLine={false} />
          <Tooltip
            content={<CustomTooltip />}
            cursor={{
              fill:
                resolvedTheme === "dark"
                  ? "rgba(255, 255, 255, 0.03)"
                  : "rgba(0, 0, 0, 0.05)",
            }}
          />
          <Bar barSize={35} dataKey="activeSprints" radius={[6, 6, 0, 0]}>
            {chartData.map((entry, index) => (
              <Cell
                fill={TEAM_COLORS[index % TEAM_COLORS.length]}
                key={`cell-${index}`}
              />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </Wrapper>
  );
};
