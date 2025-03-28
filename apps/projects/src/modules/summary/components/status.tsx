"use client";
import { Flex, Text, Wrapper, Box } from "ui";
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from "recharts";
import { useState, useEffect } from "react";
import { useStatusSummary } from "@/lib/hooks/analytics-summaries";
import type { StatusSummary } from "@/types";
import { useTerminology } from "@/hooks";

// Define colors for different statuses
const COLORS = [
  "#F06292",
  "#64B5F6",
  "#FFB74D",
  "#81C784",
  "#9575CD",
  "#4DD0E1",
  "#F8BBD0",
];

export const Status = () => {
  const { getTermDisplay } = useTerminology();
  const { data: statusSummary = [], isLoading } = useStatusSummary();
  const [chartData, setChartData] = useState<StatusSummary[]>([]);
  const [totalCount, setTotalCount] = useState(0);

  useEffect(() => {
    if (Array.isArray(statusSummary)) {
      // Use the data directly
      setChartData(statusSummary);

      // Calculate total
      const total = statusSummary.reduce(
        (sum: number, status: StatusSummary) => sum + status.count,
        0,
      );
      setTotalCount(total);
    }
  }, [statusSummary]);

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text fontSize="lg">Status overview</Text>
        <Text color="muted">See how your work is progressing.</Text>
      </Box>

      {isLoading ? (
        <Flex align="center" className="h-[300px]" justify="center">
          <Text color="muted">Loading...</Text>
        </Flex>
      ) : (
        <div className="relative">
          <ResponsiveContainer height={300} width="100%">
            <PieChart className="relative">
              <Pie
                cx="50%"
                cy="50%"
                data={chartData}
                dataKey="count"
                fill="#8884d8"
                innerRadius={60}
                labelLine={false}
                nameKey="name"
                outerRadius={100}
              >
                {chartData.map((entry, index) => (
                  <Cell
                    fill={COLORS[index % COLORS.length]}
                    key={`cell-${index}`}
                  />
                ))}
              </Pie>
              <Tooltip
                formatter={(value: number) => [`${value} items`, "Count"]}
                labelFormatter={(label: string) => label}
              />
            </PieChart>
          </ResponsiveContainer>
          <Box className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 transform text-center">
            <Text fontSize="2xl" fontWeight="bold">
              {totalCount}
            </Text>
            <Text color="muted">
              Total{" "}
              {getTermDisplay("storyTerm", {
                variant: "plural",
              })}
            </Text>
          </Box>
        </div>
      )}
    </Wrapper>
  );
};
