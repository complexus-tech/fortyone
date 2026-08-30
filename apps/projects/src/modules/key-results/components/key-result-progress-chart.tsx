"use client";

import { format } from "date-fns";
import {
  Line,
  LineChart,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Box } from "ui";
import { formatKeyResultValue } from "../utils";

export type KeyResultProgressPoint = {
  date: string;
  value: number;
};

type KeyResultMeasurementType = Parameters<typeof formatKeyResultValue>[1];

export const KeyResultProgressChart = ({
  data,
  measurementType,
  targetValue,
}: {
  data: KeyResultProgressPoint[];
  measurementType: KeyResultMeasurementType;
  targetValue: number;
}) => (
  <Box className="h-56 pt-2">
    <ResponsiveContainer height="100%" width="100%">
      <LineChart data={data} margin={{ left: -20, right: 12 }}>
        <XAxis
          axisLine={false}
          dataKey="date"
          tickFormatter={(value: string) => format(new Date(value), "MMM d")}
          tickLine={false}
        />
        <YAxis axisLine={false} tickLine={false} />
        <Tooltip
          contentStyle={{
            background: "var(--color-surface-elevated)",
            border: "1px solid var(--color-border-strong)",
            borderRadius: "0.75rem",
          }}
          formatter={(value) => [
            formatKeyResultValue(Number(value), measurementType),
            "Progress",
          ]}
          labelFormatter={(value) =>
            format(new Date(String(value)), "MMM d, yyyy")
          }
        />
        <ReferenceLine
          stroke="var(--color-text-muted)"
          strokeDasharray="4 4"
          y={targetValue}
        />
        <Line
          dataKey="value"
          dot={{ r: 3 }}
          stroke="var(--color-info)"
          strokeWidth={2}
          type="monotone"
        />
      </LineChart>
    </ResponsiveContainer>
  </Box>
);
