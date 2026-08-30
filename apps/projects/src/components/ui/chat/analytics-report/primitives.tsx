"use client";

import type { ReactNode } from "react";
import { cn } from "lib";
import { Box, Flex, Text } from "ui";
import { useTheme } from "next-themes";
import type { TooltipProps } from "recharts";
import {
  Bar,
  BarChart,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip as ChartTooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { ChartRow, Metric } from "./model";
import { formatCategoryTick, formatChartDate } from "./model";

const COMPACT_CHART_HEIGHT = 190;

export const MetricGrid = ({ metrics }: { metrics: Metric[] }) => {
  const rowCount = Math.ceil(metrics.length / 2);

  return (
    <Box className="border-border/40 grid grid-cols-2 overflow-hidden rounded-lg border dark:border-white/[0.07]">
      {metrics.map((metric, index) => {
        const rowIndex = Math.floor(index / 2);

        return (
          <Flex
            align="center"
            className={cn(
              "border-border/40 bg-surface-muted/10 min-h-10 min-w-0 gap-2 px-2.5 py-2 dark:border-white/[0.07] dark:bg-white/[0.01]",
              index % 2 === 1 && "border-l",
              rowIndex < rowCount - 1 && "border-b",
            )}
            justify="between"
            key={metric.label}
          >
            <Text className="text-foreground/55 min-w-0 truncate text-[0.95rem] font-medium dark:text-white/45">
              {metric.label}
            </Text>
            <Text className="text-foreground shrink-0 text-base font-semibold dark:text-white">
              {metric.value}
            </Text>
          </Flex>
        );
      })}
    </Box>
  );
};

export const EmptyChart = () => (
  <Box className="text-foreground/60 bg-surface-muted/30 flex h-24 items-center justify-center rounded-lg text-base dark:bg-white/[0.02] dark:text-white/55">
    No chart data available
  </Box>
);

export const ChartSection = ({
  title,
  children,
  description,
}: {
  title: string;
  children: ReactNode;
  description?: string;
}) => (
  <Box className="space-y-2">
    <Box>
      <Text className="text-foreground font-semibold dark:text-white">
        {title}
      </Text>
      {description ? (
        <Text className="mt-1 text-base leading-6" color="muted">
          {description}
        </Text>
      ) : null}
    </Box>
    {children}
  </Box>
);

export const KeyValueList = ({
  rows,
}: {
  rows: {
    key?: string;
    label: string;
    value: string | number | null | undefined;
  }[];
}) => (
  <Box className="border-border/40 divide-border/50 divide-y overflow-hidden rounded-lg border dark:divide-white/[0.07] dark:border-white/[0.07]">
    {rows.map((row) => (
      <Flex
        align="center"
        className="bg-surface-muted/10 gap-3 px-3 py-2 text-base dark:bg-white/[0.01]"
        justify="between"
        key={row.key ?? row.label}
      >
        <Text className="text-foreground/55 font-medium dark:text-white/45">
          {row.label}
        </Text>
        <Text className="text-foreground text-right font-medium dark:text-white">
          {row.value === null || row.value === undefined || row.value === ""
            ? "Not set"
            : row.value}
        </Text>
      </Flex>
    ))}
  </Box>
);

export const PillList = ({
  emptyText,
  items,
}: {
  emptyText: string;
  items: string[];
}) => {
  if (!items.length) {
    return (
      <Text className="text-foreground/60 bg-surface-muted/30 rounded-lg px-3 py-2 text-base dark:bg-white/[0.02] dark:text-white/55">
        {emptyText}
      </Text>
    );
  }

  return (
    <Flex className="flex-wrap gap-1.5">
      {items.map((item) => (
        <Text
          className="border-border/40 bg-surface-muted/20 text-foreground/90 rounded-md border px-2 py-1 text-[0.95rem] font-medium dark:border-white/[0.07] dark:bg-white/[0.02] dark:text-white/85"
          key={item}
        >
          {item}
        </Text>
      ))}
    </Flex>
  );
};

export const CompactBarChart = ({
  bars,
  data,
  maxLabelLength = 15,
  xKey,
}: {
  bars: { key: string; color: string; name?: string }[];
  data: ChartRow[];
  maxLabelLength?: number;
  xKey: string;
}) => {
  const { resolvedTheme } = useTheme();
  if (!data.length) return <EmptyChart />;

  const isDark = resolvedTheme === "dark";
  const axisColor = isDark
    ? "rgba(255, 255, 255, 0.09)"
    : "rgba(15, 23, 42, 0.12)";
  const tickColor = isDark ? "#A1A1AA" : "#64748B";

  return (
    <ResponsiveContainer height={COMPACT_CHART_HEIGHT} width="100%">
      <BarChart data={data} margin={{ top: 6, right: 4, left: -16, bottom: 0 }}>
        <XAxis
          axisLine={{ stroke: axisColor }}
          dataKey={xKey}
          minTickGap={12}
          tick={{ fill: tickColor, fontSize: 13 }}
          tickFormatter={(value) => formatCategoryTick(value, maxLabelLength)}
          tickLine={{ stroke: axisColor }}
        />
        <YAxis
          axisLine={false}
          tick={{ fill: tickColor, fontSize: 13 }}
          tickLine={false}
          width={36}
        />
        <ChartTooltip
          allowEscapeViewBox={{ x: false, y: true }}
          content={<AnalyticsChartTooltip />}
          cursor={{
            fill: isDark
              ? "rgba(255, 255, 255, 0.035)"
              : "rgba(15, 23, 42, 0.035)",
          }}
          wrapperStyle={{ outline: "none" }}
        />
        {bars.map((bar) => (
          <Bar
            activeBar={{
              fillOpacity: 0.82,
              stroke: isDark
                ? "rgba(255, 255, 255, 0.3)"
                : "rgba(15, 23, 42, 0.18)",
              strokeWidth: 1,
            }}
            dataKey={bar.key}
            fill={bar.color}
            key={bar.key}
            maxBarSize={36}
            name={bar.name}
            radius={[4, 4, 1, 1]}
          />
        ))}
      </BarChart>
    </ResponsiveContainer>
  );
};

export const CompactLineChart = ({
  data,
  xKey,
  lines,
}: {
  data: ChartRow[];
  xKey: string;
  lines: { key: string; color: string; name?: string }[];
}) => {
  const { resolvedTheme } = useTheme();
  if (!data.length) return <EmptyChart />;

  const isDark = resolvedTheme === "dark";
  const axisColor = isDark
    ? "rgba(255, 255, 255, 0.09)"
    : "rgba(15, 23, 42, 0.12)";
  const tickColor = isDark ? "#A1A1AA" : "#64748B";
  const isDateAxis = xKey.toLowerCase().includes("date");

  return (
    <ResponsiveContainer height={COMPACT_CHART_HEIGHT} width="100%">
      <LineChart
        data={data}
        margin={{ top: 6, right: 4, left: -16, bottom: 0 }}
      >
        <XAxis
          axisLine={{ stroke: axisColor }}
          dataKey={xKey}
          minTickGap={16}
          tick={{ fill: tickColor, fontSize: 13 }}
          tickFormatter={isDateAxis ? formatChartDate : formatCategoryTick}
          tickLine={{ stroke: axisColor }}
        />
        <YAxis
          axisLine={false}
          tick={{ fill: tickColor, fontSize: 13 }}
          tickLine={false}
          width={36}
        />
        <ChartTooltip
          allowEscapeViewBox={{ x: false, y: true }}
          content={
            <AnalyticsChartTooltip
              formatLabel={isDateAxis ? formatChartDate : undefined}
            />
          }
          cursor={{
            stroke: isDark
              ? "rgba(255, 255, 255, 0.16)"
              : "rgba(15, 23, 42, 0.18)",
            strokeDasharray: "3 4",
            strokeWidth: 1,
          }}
          wrapperStyle={{ outline: "none" }}
        />
        {lines.map((line) => (
          <Line
            activeDot={{
              fill: line.color,
              r: 6,
              stroke: isDark ? "#18181B" : "#FFFFFF",
              strokeWidth: 3,
            }}
            dataKey={line.key}
            dot={false}
            key={line.key}
            name={line.name}
            stroke={line.color}
            strokeWidth={2}
            type="monotone"
          />
        ))}
      </LineChart>
    </ResponsiveContainer>
  );
};

const AnalyticsChartTooltip = ({
  active,
  formatLabel,
  label,
  payload,
}: TooltipProps<number, string> & {
  formatLabel?: (value: unknown) => string;
}) => {
  if (!active || !payload?.length) return null;

  return (
    <Box className="border-border/40 bg-surface-elevated/85 text-foreground z-50 min-w-40 rounded-2xl border-[0.5px] p-4 shadow-lg shadow-black/5 backdrop-blur-xl dark:border-white/[0.08] dark:shadow-black/20">
      <Text className="text-base font-semibold">
        {formatLabel ? formatLabel(label) : String(label ?? "")}
      </Text>
      <Box className="mt-2 space-y-1.5">
        {payload.map((entry) => (
          <Flex
            align="center"
            className="gap-2 text-[0.95rem] font-medium"
            key={String(entry.dataKey)}
          >
            <span
              aria-hidden
              className="size-2.5 shrink-0 rounded-full"
              style={{ backgroundColor: entry.color }}
            />
            <Text className="text-foreground/75 dark:text-white/70">
              {entry.name}:{" "}
              <span className="text-foreground dark:text-white">
                {entry.value}
              </span>
            </Text>
          </Flex>
        ))}
      </Box>
    </Box>
  );
};

export const HealthAndProgressSection = ({
  description,
  healthRows,
  progressRows,
  title,
}: {
  description: string;
  healthRows: {
    key?: string;
    label: string;
    value: string | number | null | undefined;
  }[];
  progressRows: {
    key?: string;
    label: string;
    value: string | number | null | undefined;
  }[];
  title: string;
}) => (
  <ChartSection description={description} title={title}>
    <Box className="space-y-3">
      <KeyValueList rows={healthRows} />
      {progressRows.length ? (
        <Box className="space-y-2">
          <Text className="text-foreground/55 text-base font-medium dark:text-white/45">
            Progress
          </Text>
          <KeyValueList rows={progressRows} />
        </Box>
      ) : null}
    </Box>
  </ChartSection>
);

export const RiskList = ({ risks }: { risks: ChartRow[] }) => {
  if (!risks.length) {
    return (
      <Text className="text-text-muted border-border/50 border-y py-2 text-base dark:border-white/[0.07]">
        No active risks found.
      </Text>
    );
  }

  return (
    <Box className="border-border/50 divide-border/50 divide-y border-y dark:divide-white/[0.07] dark:border-white/[0.07]">
      {risks.slice(0, 5).map((risk) => {
        const severity = String(risk.severity ?? "low");
        let tone = "bg-info";
        if (severity === "high") tone = "bg-danger";
        if (severity === "medium") tone = "bg-warning";

        return (
          <Flex
            align="center"
            className="min-w-0 gap-3 py-2 text-base"
            justify="between"
            key={String(
              risk.kind ??
                `${String(risk.title ?? "Risk")}-${String(risk.severity ?? "low")}`,
            )}
          >
            <span className={`size-2.5 shrink-0 rounded-[2px] ${tone}`} />
            <Text className="min-w-0 flex-1 truncate">
              {String(risk.title ?? "Risk")}
            </Text>
            <Text className="text-text-muted shrink-0">
              {Number(risk.count ?? 0)}
            </Text>
          </Flex>
        );
      })}
    </Box>
  );
};
