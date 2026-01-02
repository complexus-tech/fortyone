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
import { useWorkspaceOverview } from "../hooks/workspace-overview";
import type { VelocityTrendPoint } from "../types";
import { useAppliedFilters } from "../hooks/filters";
import { VelocityTrendSkeleton } from "./velocity-trend-skeleton";

type ChartDataItem = {
  period: string;
  velocity: number;
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
      <Text className="mt-1 pl-0.5">{payload[0].value} velocity</Text>
    </Box>
  );
};

export const VelocityTrend = () => {
  const { resolvedTheme } = useTheme();
  const filters = useAppliedFilters();
  const { data: overview, isPending } = useWorkspaceOverview(filters);
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);

  useEffect(() => {
    if (overview?.velocityTrend.length) {
      const formattedData = overview.velocityTrend.map(
        (item: VelocityTrendPoint) => ({
          period: item.period,
          velocity: item.velocity,
        }),
      );
      setChartData(formattedData);
    }
  }, [overview]);

  if (isPending) {
    return <VelocityTrendSkeleton />;
  }

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Velocity trend
        </Text>
        <Text color="muted">Team velocity over time.</Text>
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
            dataKey="period"
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
            dataKey="velocity"
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
