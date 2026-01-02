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
} from "recharts";
import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { useTeamPerformance } from "../hooks/team-performance";
import type { VelocityByTeamItem } from "../types";
import { useAppliedFilters } from "../hooks/filters";
import { TeamVelocitySkeleton } from "./team-velocity-skeleton";

type ChartDataItem = {
  teamName: string;
  average: number;
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
    <Box className="z-50 min-w-28 rounded-[0.6rem] border border-border bg-surface-elevated/80 px-3 py-3 text-[0.95rem] font-medium text-foreground backdrop-blur">
      <Flex align="center" gap={2}>
        {label}
      </Flex>
      <Text className="mt-1 pl-0.5">Avg velocity: {payload[0].value}</Text>
    </Box>
  );
};

export const TeamVelocity = () => {
  const { resolvedTheme } = useTheme();
  const filters = useAppliedFilters();
  const { data: teamPerformance, isPending } = useTeamPerformance(filters);
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);

  useEffect(() => {
    if (teamPerformance?.velocityByTeam.length) {
      const formattedData = teamPerformance.velocityByTeam.map(
        (item: VelocityByTeamItem) => ({
          teamName: item.teamName,
          average: item.average,
        }),
      );
      setChartData(formattedData);
    }
  }, [teamPerformance]);

  if (isPending) {
    return <TeamVelocitySkeleton />;
  }

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Team velocity
        </Text>
        <Text color="muted">Average velocity comparison by team.</Text>
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
          <Bar
            barSize={35}
            dataKey="average"
            fill="#6366F1"
            radius={[6, 6, 0, 0]}
          />
        </BarChart>
      </ResponsiveContainer>
    </Wrapper>
  );
};
