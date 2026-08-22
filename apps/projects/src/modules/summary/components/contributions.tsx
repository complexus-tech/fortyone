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
import { useMemo } from "react";
import type { Contribution } from "@/types";
import { useContributions } from "@/lib/hooks/contributions";
import { useSummaryDateFilters } from "@/modules/summary/hooks/summary-date-filters";
import { ContributionsSkeleton } from "./contributions-skeleton";
import {
  SUMMARY_CHART_CURSOR,
  SUMMARY_CHART_GRID,
  SUMMARY_CHART_PRIMARY,
  SUMMARY_CHART_SURFACE,
} from "./chart-colors";

const CustomTooltip = ({
  active,
  payload,
  label,
}: TooltipProps<number, string>) => {
  if (!active || !payload?.length) {
    return null;
  }

  return (
    <Box className="border-border/60 bg-surface-elevated/80 text-foreground z-50 min-w-28 rounded-lg border-[0.5px] px-3 py-3 text-[0.95rem] font-medium backdrop-blur">
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
  const filters = useSummaryDateFilters();
  const { data: contributions = [], isLoading } = useContributions(filters);
  const chartData = useMemo(() => {
    return contributions.map((item: Contribution) => ({
      date: formatDate(item.date),
      count: item.contributions,
    }));
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
            <linearGradient
              id="summaryContributionsGradient"
              x1="0"
              x2="0"
              y1="0"
              y2="1"
            >
              <stop
                offset="5%"
                stopColor={SUMMARY_CHART_PRIMARY}
                stopOpacity={0.28}
              />
              <stop
                offset="95%"
                stopColor={SUMMARY_CHART_PRIMARY}
                stopOpacity={0.02}
              />
            </linearGradient>
          </defs>
          <CartesianGrid
            stroke={SUMMARY_CHART_GRID}
            strokeDasharray="3 3"
            vertical={false}
          />
          <XAxis
            axisLine={{ stroke: SUMMARY_CHART_GRID }}
            dataKey="date"
            tick={{ fontSize: 12 }}
          />
          <YAxis axisLine={false} tick={{ fontSize: 12 }} tickLine={false} />
          <Tooltip
            content={<CustomTooltip />}
            cursor={{ fill: SUMMARY_CHART_CURSOR }}
          />
          <Area
            activeDot={{
              r: 4,
              fill: SUMMARY_CHART_PRIMARY,
              strokeWidth: 2,
              stroke: SUMMARY_CHART_SURFACE,
            }}
            dataKey="count"
            dot={{
              r: 2,
              fill: SUMMARY_CHART_PRIMARY,
              strokeWidth: 2,
              stroke: SUMMARY_CHART_SURFACE,
            }}
            fill="url(#summaryContributionsGradient)"
            stroke={SUMMARY_CHART_PRIMARY}
            strokeOpacity={0.82}
            strokeWidth={2}
            type="monotone"
          />
        </AreaChart>
      </ResponsiveContainer>
    </Wrapper>
  );
};
