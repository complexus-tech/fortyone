import type { TooltipProps } from "recharts";
import { useTheme } from "next-themes";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Line,
  LineChart,
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
import type {
  ChartBreakdownRow,
  CompletionTrendChartRow,
  MemberWorkloadChartRow,
  ProgressChartRow,
} from "./model";
import { EmptyState } from "./primitives";

type ProgressSegmentShapeProps = {
  fill?: string;
  height?: number;
  payload?: ProgressChartRow;
  segment: "completed" | "remaining";
  width?: number;
  x?: number;
  y?: number;
};

const ChartTooltip = ({
  active,
  label,
  payload,
}: TooltipProps<number, string>) => {
  if (!active || !payload?.length) {
    return null;
  }

  return (
    <Box className="border-border/60 bg-surface-elevated z-50 min-w-36 rounded-lg border-[0.5px] px-3 py-3 text-[0.95rem] shadow-lg">
      <Text className="mb-2 font-medium">{label}</Text>
      <Box className="space-y-1">
        {payload.map((entry) => (
          <Flex align="center" className="gap-2" key={String(entry.dataKey)}>
            <Box
              className="size-2.5 rounded-sm"
              style={{ backgroundColor: entry.color }}
            />
            <Text color="muted">
              {entry.name ?? entry.dataKey}: {formatNumber(Number(entry.value))}
            </Text>
          </Flex>
        ))}
      </Box>
    </Box>
  );
};

export const DeliveryChart = ({
  data,
}: {
  data: CompletionTrendChartRow[];
}) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);

  if (!data.length) {
    return <EmptyState>No completion trend data is available.</EmptyState>;
  }

  return (
    <ResponsiveContainer height={300} width="100%">
      <LineChart
        data={data}
        margin={{ top: 8, right: 12, left: -18, bottom: 4 }}
      >
        <CartesianGrid
          stroke={gridStroke}
          strokeDasharray="3 3"
          vertical={false}
        />
        <XAxis dataKey="date" tick={{ fontSize: 12 }} tickLine={false} />
        <YAxis axisLine={false} tick={{ fontSize: 12 }} tickLine={false} />
        <Tooltip
          content={<ChartTooltip />}
          cursor={{ stroke: getCursorFill(resolvedTheme), strokeWidth: 1 }}
        />
        <Line
          dataKey="total"
          dot={false}
          name="Total"
          stroke={chartPalette.primary}
          strokeWidth={2.5}
          type="monotone"
        />
        <Line
          dataKey="completed"
          dot={false}
          name="Completed"
          stroke={chartPalette.success}
          strokeWidth={2.5}
          type="monotone"
        />
      </LineChart>
    </ResponsiveContainer>
  );
};

export const WorkloadChart = ({ data }: { data: MemberWorkloadChartRow[] }) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);

  if (!data.length) {
    return <EmptyState>No member workload chart data is available.</EmptyState>;
  }

  return (
    <ResponsiveContainer height={300} width="100%">
      <BarChart
        data={data}
        margin={{ top: 8, right: 12, left: -18, bottom: 4 }}
      >
        <CartesianGrid
          stroke={gridStroke}
          strokeDasharray="3 3"
          vertical={false}
        />
        <XAxis dataKey="name" tick={{ fontSize: 12 }} tickLine={false} />
        <YAxis axisLine={false} tick={{ fontSize: 12 }} tickLine={false} />
        <Tooltip
          content={<ChartTooltip />}
          cursor={{ fill: getCursorFill(resolvedTheme) }}
        />
        <Bar
          dataKey="open"
          fill={chartPalette.primary}
          name="Open"
          radius={[4, 4, 0, 0]}
        />
        <Bar
          dataKey="overdue"
          fill={chartPalette.danger}
          name="Overdue"
          radius={[4, 4, 0, 0]}
        />
      </BarChart>
    </ResponsiveContainer>
  );
};

export const HorizontalBreakdownChart = ({
  color,
  data,
  emptyText,
  height = 300,
  labelWidth = 112,
}: {
  color: string;
  data: ChartBreakdownRow[];
  emptyText: string;
  height?: number;
  labelWidth?: number;
}) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);
  const chartData = data.map((item) => ({
    color: item.color ?? color,
    name: item.label,
    value: item.value,
  }));

  if (!chartData.length) {
    return <EmptyState>{emptyText}</EmptyState>;
  }

  return (
    <ResponsiveContainer height={height} width="100%">
      <BarChart
        data={chartData}
        layout="vertical"
        margin={{ top: 4, right: 20, left: -18, bottom: 4 }}
      >
        <CartesianGrid
          horizontal={false}
          stroke={gridStroke}
          strokeDasharray="3 3"
        />
        <XAxis
          axisLine={false}
          tick={{ fontSize: 12 }}
          tickLine={false}
          type="number"
        />
        <YAxis
          axisLine={false}
          dataKey="name"
          tick={{ fontSize: 12 }}
          tickFormatter={(value: string) => truncateChartLabel(value, 20)}
          tickLine={false}
          type="category"
          width={labelWidth}
        />
        <Tooltip
          content={<ChartTooltip />}
          cursor={{ fill: getCursorFill(resolvedTheme) }}
        />
        <Bar dataKey="value" name="Count" radius={[0, 4, 4, 0]}>
          {chartData.map((entry) => (
            <Cell fill={entry.color} key={entry.name} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
};

export const ProgressComparisonChart = ({
  data,
  emptyText,
  height = 220,
}: {
  data: ProgressChartRow[];
  emptyText: string;
  height?: number;
}) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);

  if (!data.length) {
    return <EmptyState>{emptyText}</EmptyState>;
  }

  return (
    <ResponsiveContainer height={height} width="100%">
      <BarChart
        data={data}
        layout="vertical"
        margin={{ top: 4, right: 20, left: -18, bottom: 4 }}
      >
        <CartesianGrid
          horizontal={false}
          stroke={gridStroke}
          strokeDasharray="3 3"
        />
        <XAxis
          axisLine={false}
          tick={{ fontSize: 12 }}
          tickLine={false}
          type="number"
        />
        <YAxis
          axisLine={false}
          dataKey="label"
          tick={{ fontSize: 12 }}
          tickFormatter={(value: string) => truncateChartLabel(value, 20)}
          tickLine={false}
          type="category"
          width={120}
        />
        <Tooltip
          content={<ChartTooltip />}
          cursor={{ fill: getCursorFill(resolvedTheme) }}
        />
        <Bar
          dataKey="completed"
          fill={chartPalette.success}
          name="Completed"
          shape={
            <ProgressSegmentShape
              fill={chartPalette.success}
              segment="completed"
            />
          }
          stackId="progress"
        />
        <Bar
          dataKey="remaining"
          fill={chartPalette.primary}
          name="Remaining"
          shape={
            <ProgressSegmentShape
              fill={chartPalette.primary}
              segment="remaining"
            />
          }
          stackId="progress"
        />
      </BarChart>
    </ResponsiveContainer>
  );
};

const ProgressSegmentShape = ({
  fill = chartPalette.primary,
  height = 0,
  payload,
  segment,
  width = 0,
  x = 0,
  y = 0,
}: ProgressSegmentShapeProps) => {
  if (width <= 0 || height <= 0) {
    return null;
  }

  const shouldRoundRight =
    segment === "remaining" || (payload?.remaining ?? 0) <= 0;

  if (!shouldRoundRight) {
    return <rect fill={fill} height={height} width={width} x={x} y={y} />;
  }

  const radius = Math.min(4, width / 2, height / 2);
  const right = x + width;
  const bottom = y + height;

  return (
    <path
      d={[
        `M ${x} ${y}`,
        `H ${right - radius}`,
        `Q ${right} ${y} ${right} ${y + radius}`,
        `V ${bottom - radius}`,
        `Q ${right} ${bottom} ${right - radius} ${bottom}`,
        `H ${x}`,
        "Z",
      ].join(" ")}
      fill={fill}
    />
  );
};
