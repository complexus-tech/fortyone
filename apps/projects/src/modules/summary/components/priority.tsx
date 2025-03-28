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
import { usePrioritySummary } from "@/lib/hooks/analytics-summaries";
import type { PrioritySummary } from "@/types";
import { PriorityIcon } from "@/components/ui";
import type { StoryPriority } from "@/modules/stories/types";
import { useTerminology } from "@/hooks";

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
    <Box className="z-50 min-w-28 rounded-[0.6rem] border border-gray-100 bg-white/80 px-3 py-3 text-[0.95rem] font-medium text-gray backdrop-blur dark:border-dark-50 dark:bg-dark-200 dark:text-gray-200">
      <Flex align="center" gap={2}>
        <PriorityIcon priority={label as StoryPriority} />
        {label}
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

export const Priority = () => {
  const { resolvedTheme } = useTheme();
  const { data: prioritySummary = [], isLoading } = usePrioritySummary();
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);

  useEffect(() => {
    if (prioritySummary.length > 0) {
      const formattedData = prioritySummary.map((item: PrioritySummary) => ({
        priority: item.priority,
        count: item.count,
      }));
      setChartData(formattedData);
    }
  }, [prioritySummary]);

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Priority breakdown
        </Text>
        <Text color="muted">
          Get a holistic view of how your work is prioritized.
        </Text>
      </Box>

      {isLoading ? (
        <Flex align="center" className="h-[300px]" justify="center">
          <Text color="muted">Loading...</Text>
        </Flex>
      ) : (
        <ResponsiveContainer height={300} width="100%">
          <BarChart
            data={chartData}
            margin={{ top: 20, right: 10, left: -20, bottom: 5 }}
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
      )}
    </Wrapper>
  );
};
