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
import type { Contribution } from "@/types";
import { useContributions } from "@/lib/hooks/contributions";

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
    <Box className="z-50 min-w-28 rounded-[0.6rem] border border-gray-100 bg-white/80 px-3 py-3 text-[0.95rem] font-medium text-gray backdrop-blur dark:border-dark-50 dark:bg-dark-200 dark:text-gray-200">
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
  const { data: contributions = [], isLoading } = useContributions();
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

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Daily contributions
        </Text>
        <Text color="muted">Your contributions in the last 7 days.</Text>
      </Box>

      {isLoading ? (
        <Flex align="center" className="h-[220px]" justify="center">
          <Text color="muted">Loading...</Text>
        </Flex>
      ) : (
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
            <Bar
              barSize={20}
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
