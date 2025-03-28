"use client";
import { Flex, Text, Wrapper } from "ui";
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Legend,
  Tooltip,
} from "recharts";
import { useState, useEffect } from "react";
import { useStatusSummary } from "@/lib/hooks/analytics-summaries";
import type { StatusSummary } from "@/types";

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

// Create a legend item component outside the render function
const LegendItem = ({ value }: { value: string }) => (
  <span style={{ color: "#666", fontSize: "12px" }}>{value}</span>
);

type PieChartLabelProps = {
  cx: number;
  cy: number;
  midAngle: number;
  innerRadius: number;
  outerRadius: number;
  percent: number;
};

export const Status = () => {
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

  const renderCustomizedLabel = ({
    cx,
    cy,
    midAngle,
    innerRadius,
    outerRadius,
    percent,
  }: PieChartLabelProps) => {
    const RADIAN = Math.PI / 180;
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    return percent > 0.05 ? (
      <text
        dominantBaseline="central"
        fill="white"
        fontSize={12}
        textAnchor={x > cx ? "start" : "end"}
        x={x}
        y={y}
      >
        {`${(percent * 100).toFixed(0)}%`}
      </text>
    ) : null;
  };

  // Memoize the legend formatter function to prevent new component creation on each render
  const legendFormatter = (value: string) => <LegendItem value={value} />;

  return (
    <Wrapper>
      <Flex align="center" className="mb-5" justify="between">
        <Text fontSize="lg">Status overview</Text>
      </Flex>

      {isLoading ? (
        <Flex align="center" className="h-[300px]" justify="center">
          <Text color="muted">Loading...</Text>
        </Flex>
      ) : (
        <div className="relative">
          <ResponsiveContainer height={300} width="100%">
            <PieChart>
              <Pie
                cx="50%"
                cy="50%"
                data={chartData}
                dataKey="count"
                fill="#8884d8"
                innerRadius={60}
                label={renderCustomizedLabel}
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
              <Legend
                align="right"
                formatter={legendFormatter}
                iconSize={10}
                iconType="circle"
                layout="vertical"
                verticalAlign="middle"
              />
            </PieChart>
          </ResponsiveContainer>
          <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 transform text-center">
            <Text fontSize="xl" fontWeight="bold">
              {totalCount}
            </Text>
            <Text color="muted" fontSize="sm">
              Total issues
            </Text>
          </div>
        </div>
      )}
    </Wrapper>
  );
};
