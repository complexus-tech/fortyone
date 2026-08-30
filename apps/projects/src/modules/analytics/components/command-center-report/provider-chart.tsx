import type { TooltipProps } from "recharts";
import { useTheme } from "next-themes";
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Box, Flex, Text } from "ui";
import {
  chartPalette,
  formatNumber,
  getCursorFill,
  getGridStroke,
  truncateChartLabel,
} from "./model";
import type { ProviderChartRow } from "./model";
import { EmptyState } from "./primitives";

const ProviderTooltip = ({ active, payload }: TooltipProps<number, string>) => {
  const row = payload?.[0]?.payload as ProviderChartRow | undefined;

  if (!active || !row) {
    return null;
  }

  const items = [
    {
      color: chartPalette.success,
      count: row.accepted,
      label: "Accepted",
      share: row.acceptedShare,
    },
    {
      color: chartPalette.warning,
      count: row.pending,
      label: "Pending",
      share: row.pendingShare,
    },
    {
      color: chartPalette.danger,
      count: row.stale,
      label: "Stale",
      share: row.staleShare,
    },
  ];

  return (
    <Box className="border-border/60 bg-surface-elevated z-50 min-w-44 rounded-lg border-[0.5px] px-3 py-3 text-[0.95rem] shadow-lg">
      <Flex align="center" className="mb-2 gap-2" justify="between">
        <Text className="font-medium">{row.provider}</Text>
        <Text color="muted">{formatNumber(row.total)} total</Text>
      </Flex>
      <Box className="space-y-1">
        {items.map((item) => (
          <Flex align="center" className="gap-2" key={item.label}>
            <Box
              className="size-2.5 rounded-sm"
              style={{ backgroundColor: item.color }}
            />
            <Text color="muted">
              {item.label}: {formatNumber(item.count)} ({Math.round(item.share)}
              %)
            </Text>
          </Flex>
        ))}
      </Box>
    </Box>
  );
};

export const ProviderChart = ({ data }: { data: ProviderChartRow[] }) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);

  if (!data.length) {
    return <EmptyState>No request source data is available.</EmptyState>;
  }

  return (
    <ResponsiveContainer
      height={Math.max(300, data.length * 42 + 90)}
      width="100%"
    >
      <BarChart
        data={data}
        layout="vertical"
        margin={{ top: 8, right: 18, left: -12, bottom: 4 }}
      >
        <CartesianGrid
          horizontal={false}
          stroke={gridStroke}
          strokeDasharray="3 3"
        />
        <XAxis
          axisLine={false}
          domain={[0, 100]}
          tick={{ fontSize: 12 }}
          tickFormatter={(value: number) => `${value}%`}
          tickLine={false}
          type="number"
        />
        <YAxis
          axisLine={false}
          dataKey="provider"
          tick={{ fontSize: 12 }}
          tickFormatter={(value: string) => truncateChartLabel(value, 18)}
          tickLine={false}
          type="category"
          width={96}
        />
        <Tooltip
          content={<ProviderTooltip />}
          cursor={{ fill: getCursorFill(resolvedTheme) }}
        />
        <Bar
          dataKey="acceptedShare"
          fill={chartPalette.success}
          name="Accepted"
          radius={[0, 0, 0, 0]}
          stackId="requests"
        />
        <Bar
          dataKey="pendingShare"
          fill={chartPalette.warning}
          name="Pending"
          radius={[0, 0, 0, 0]}
          stackId="requests"
        />
        <Bar
          dataKey="staleShare"
          fill={chartPalette.danger}
          name="Stale"
          radius={[0, 4, 4, 0]}
          stackId="requests"
        />
      </BarChart>
    </ResponsiveContainer>
  );
};

export const ProviderLegend = () => {
  const items = [
    { color: chartPalette.success, label: "Accepted" },
    { color: chartPalette.warning, label: "Pending" },
    { color: chartPalette.danger, label: "Stale" },
  ];

  return (
    <Flex align="center" className="mt-3 flex-wrap gap-x-5 gap-y-2">
      {items.map((item) => (
        <Flex align="center" className="gap-2" key={item.label}>
          <Box
            className="size-2.5 rounded-full"
            style={{ backgroundColor: item.color }}
          />
          <Text className="text-foreground/80 text-[0.92rem]">
            {item.label}
          </Text>
        </Flex>
      ))}
    </Flex>
  );
};
