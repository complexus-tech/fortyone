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
import { PriorityIcon } from "@/components/ui";
import type { StoryPriority } from "@/modules/stories/types";
import { useTerminology } from "@/hooks";
import type { PriorityDistributionItem } from "../types";
import { useStoryAnalytics } from "../hooks/story-analytics";
import { useAppliedFilters } from "../hooks/filters";
import { PriorityDistributionSkeleton } from "./priority-distribution-skeleton";

type ChartDataItem = {
  priority: string;
  count: number;
};

const CustomTooltip = ({
  active,
  payload,
  label,
}: TooltipProps<number, string>) => {
  const { getTermDisplay } = useTerminology();
  if (!active || !payload?.length) {
    return null;
  }

  return (
    <Box className="z-50 min-w-28 rounded-[0.6rem] border border-border bg-surface-elevated/80 px-3 py-3 text-[0.95rem] font-medium text-foreground backdrop-blur">
      <Text className="mb-1 pl-0.5">
        {payload[0].value}{" "}
        {getTermDisplay("storyTerm", {
          variant: payload[0].value === 1 ? "singular" : "plural",
        })}
      </Text>
      <Flex align="center" gap={2}>
        <PriorityIcon priority={label as StoryPriority} />
        {label}
      </Flex>
    </Box>
  );
};

export const PriorityDistribution = () => {
  const { resolvedTheme } = useTheme();
  const filters = useAppliedFilters();
  const { data: storyAnalytics, isPending } = useStoryAnalytics(filters);
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);

  useEffect(() => {
    if (storyAnalytics?.priorityDistribution.length) {
      const formattedData = storyAnalytics.priorityDistribution.map(
        (item: PriorityDistributionItem) => ({
          priority: item.priority,
          count: item.count,
        }),
      );
      setChartData(formattedData);
    }
  }, [storyAnalytics]);

  if (isPending) {
    return <PriorityDistributionSkeleton />;
  }

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Priority distribution
        </Text>
        <Text color="muted">Stories broken down by priority level.</Text>
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
            dataKey="priority"
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
            dataKey="count"
            fill="#6366F1"
            radius={[6, 6, 0, 0]}
          />
        </BarChart>
      </ResponsiveContainer>
    </Wrapper>
  );
};
