"use client";
import { Box, Flex, Text, Wrapper } from "ui";
import type { TooltipProps } from "recharts";
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
} from "recharts";
import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import type { Contribution } from "@/types";
import { useContributions } from "@/lib/hooks/contributions";
import { useSummaryDateFilters } from "@/modules/summary/hooks/summary-date-filters";
import { ContributionsSkeleton } from "./contributions-skeleton";

type ChartDataItem = {
  date: string;
  count: number;
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
    <Box className="border-border bg-surface-elevated/80 text-foreground z-50 min-w-28 rounded-[0.6rem] border px-3 py-3 text-[0.95rem] font-medium backdrop-blur">
      <Flex align="center" gap={2}>
        {label}
      </Flex>
      <Text className="mt-1 pl-0.5">{payload[0].value} contributions</Text>
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

export const Contributions = () => {
  const { resolvedTheme } = useTheme();
  const filters = useSummaryDateFilters();
  const { data: contributions = [], isLoading } = useContributions(filters);
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);

  useEffect(() => {
    if (contributions.length > 0) {
      const formattedData = contributions.map((item: Contribution) => ({
        date: formatDate(item.date),
        count: item.contributions,
      }));
      setChartData(formattedData);
    }
  }, [contributions]);

  if (isLoading) {
    return <ContributionsSkeleton />;
  }

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Daily contributions
        </Text>
        <Text color="muted">Your contribution activity.</Text>
      </Box>

      <ResponsiveContainer height={220} width="100%">
        <AreaChart
          data={chartData}
          margin={{ top: 20, right: 10, left: -35, bottom: 0 }}
        >
          <defs>
            <linearGradient id="colorGradient" x1="0" x2="0" y1="0" y2="1">
              <stop offset="5%" stopColor="#6366F1" stopOpacity={0.4} />
              <stop offset="95%" stopColor="#6366F1" stopOpacity={0.05} />
            </linearGradient>
          </defs>
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
              fill:
                resolvedTheme === "dark"
                  ? "rgba(255, 255, 255, 0.03)"
                  : "rgba(0, 0, 0, 0.05)",
            }}
          />
          <Area
            activeDot={{
              r: 4,
              fill: "#6366F1",
              strokeWidth: 2,
              stroke:
                resolvedTheme === "dark"
                  ? "rgba(255, 255, 255, 0.8)"
                  : "#6366F1",
            }}
            dataKey="count"
            dot={{
              r: 2,
              fill: "#6366F1",
              strokeWidth: 2,
              stroke:
                resolvedTheme === "dark"
                  ? "rgba(255, 255, 255, 0.8)"
                  : "#6366F1",
            }}
            fill="url(#colorGradient)"
            stroke="#6366F1"
            strokeWidth={2}
            type="monotone"
          />
        </AreaChart>
      </ResponsiveContainer>
    </Wrapper>
  );
};
