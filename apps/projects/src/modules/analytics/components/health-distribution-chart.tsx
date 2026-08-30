"use client";

import type { TooltipProps } from "recharts";
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";
import { Box, Flex, Text, Wrapper } from "ui";

type DistributionItem = {
  count: number;
  status: string;
};

type HealthDistributionChartProps = {
  data: DistributionItem[];
  description: string;
  heading: string;
  itemLabel: string;
  statusColors: Record<string, string>;
  totalLabel: string;
};

const FALLBACK_COLORS = [
  "#6366F1",
  "#22c55e",
  "#eab308",
  "#EA6060",
  "#06b6d4",
  "#f43f5e",
  "#8b5cf6",
  "#f97316",
];

const HealthDistributionTooltip = ({
  active,
  itemLabel,
  payload,
}: TooltipProps<number, string> & { itemLabel: string }) => {
  if (!active || !payload?.length) {
    return null;
  }

  return (
    <Box className="border-border/60 bg-surface-elevated/80 text-foreground relative z-50 min-w-28 rounded-lg border-[0.5px] px-3 py-3 text-[0.95rem] font-medium backdrop-blur">
      <Flex align="center" gap={2}>
        {payload[0].name}
      </Flex>
      <Text className="mt-1 pl-0.5">
        {payload[0].value} {itemLabel}
      </Text>
    </Box>
  );
};

export const HealthDistributionChart = ({
  data,
  description,
  heading,
  itemLabel,
  statusColors,
  totalLabel,
}: HealthDistributionChartProps) => {
  const totalCount = data.reduce((sum, item) => sum + item.count, 0);
  const getColor = (status: string, index: number) =>
    statusColors[status] ?? FALLBACK_COLORS[index % FALLBACK_COLORS.length];

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text fontSize="lg">{heading}</Text>
        <Text color="muted">{description}</Text>
      </Box>

      <Box>
        <Box className="relative isolate">
          <Box className="absolute top-1/2 left-1/2 z-0 -translate-x-1/2 -translate-y-1/2 transform text-center">
            <Text fontSize="3xl">{totalCount}</Text>
            <Text color="muted">{totalLabel}</Text>
          </Box>
          <ResponsiveContainer height={160} width="100%">
            <PieChart className="relative">
              <Pie
                cornerRadius={4}
                cx="50%"
                cy="50%"
                data={data}
                dataKey="count"
                fill="#8884d8"
                innerRadius={60}
                labelLine={false}
                nameKey="status"
                outerRadius={80}
                paddingAngle={2}
                stroke="none"
              >
                {data.map((entry, index) => (
                  <Cell
                    fill={getColor(entry.status, index)}
                    key={`cell-${entry.status}`}
                    stroke="none"
                  />
                ))}
              </Pie>
              <Tooltip
                content={<HealthDistributionTooltip itemLabel={itemLabel} />}
              />
            </PieChart>
          </ResponsiveContainer>
        </Box>
        <Flex className="line-clamp-2 h-[60px] pt-3" gap={3} wrap>
          {data.map((entry, index) => (
            <Flex align="center" gap={1} key={entry.status}>
              <Box
                className="size-4 rounded"
                style={{ backgroundColor: getColor(entry.status, index) }}
              />
              <Text>{entry.status}</Text>
            </Flex>
          ))}
        </Flex>
      </Box>
    </Wrapper>
  );
};
