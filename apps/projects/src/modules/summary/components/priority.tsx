"use client";
import { Box, Flex, Text, Wrapper } from "ui";
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

type ChartDataItem = {
  priority: string;
  count: number;
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
              cursor={{
                fill:
                  resolvedTheme === "dark"
                    ? "rgba(255, 255, 255, 0.03)"
                    : "rgba(0, 0, 0, 0.05)",
              }}
              formatter={(value: number) => [`${value} items`, "Count"]}
              labelFormatter={(label: string) => `${label} Priority`}
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
