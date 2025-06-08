"use client";
import { Box, Flex, Text, Wrapper } from "ui";
import type { TooltipProps } from "recharts";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { useTeamPerformance } from "../hooks/team-performance";
import type { TeamWorkloadItem } from "../types";
import { TeamWorkloadSkeleton } from "./team-workload-skeleton";

type ChartDataItem = {
  teamName: string;
  assigned: number;
  completed: number;
  capacity: number;
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
    <Box className="z-50 min-w-32 rounded-[0.6rem] border border-gray-100 bg-white/80 px-3 py-3 text-[0.95rem] font-medium text-gray backdrop-blur dark:border-dark-50 dark:bg-dark-200 dark:text-gray-200">
      <Text className="mb-2 font-medium">{label}</Text>
      {payload.map((entry, index) => (
        <Flex align="center" gap={2} key={index}>
          <Box
            className="h-3 w-3 rounded"
            style={{ backgroundColor: entry.color }}
          />
          <Text className="capitalize">
            {entry.dataKey}: {entry.value}
          </Text>
        </Flex>
      ))}
    </Box>
  );
};

export const TeamWorkload = () => {
  const { resolvedTheme } = useTheme();
  const { data: teamPerformance, isPending } = useTeamPerformance();
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);

  useEffect(() => {
    if (teamPerformance?.teamWorkload.length) {
      const formattedData = teamPerformance.teamWorkload.map(
        (item: TeamWorkloadItem) => ({
          teamName: item.teamName,
          assigned: item.assigned,
          completed: item.completed,
          capacity: item.capacity,
        }),
      );
      setChartData(formattedData);
    }
  }, [teamPerformance]);

  if (isPending) {
    return <TeamWorkloadSkeleton />;
  }

  return (
    <Wrapper className="my-4">
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Team workload
        </Text>
        <Text color="muted">Work distribution and capacity by team.</Text>
      </Box>

      <ResponsiveContainer height={280} width="100%">
        <BarChart
          data={chartData}
          margin={{ top: 20, right: 30, left: 20, bottom: 60 }}
        >
          <CartesianGrid
            stroke={resolvedTheme === "dark" ? "#222" : "#E0E0E0"}
            strokeDasharray="3 3"
            vertical={false}
          />
          <XAxis
            angle={-45}
            axisLine={{
              stroke: resolvedTheme === "dark" ? "#222" : "#E0E0E0",
            }}
            dataKey="teamName"
            height={60}
            textAnchor="end"
            tick={{ fontSize: 12 }}
          />
          <YAxis axisLine={false} tick={{ fontSize: 12 }} tickLine={false} />
          <Tooltip content={<CustomTooltip />} />
          <Legend />
          <Bar
            dataKey="completed"
            fill="#22c55e"
            name="Completed"
            radius={[0, 0, 0, 0]}
            stackId="workload"
          />
          <Bar
            dataKey="assigned"
            fill="#6366F1"
            name="Assigned"
            radius={[0, 0, 0, 0]}
            stackId="workload"
          />
          <Bar
            dataKey="capacity"
            fill="#eab308"
            name="Capacity"
            radius={[4, 4, 0, 0]}
            stackId="capacity"
          />
        </BarChart>
      </ResponsiveContainer>
    </Wrapper>
  );
};
