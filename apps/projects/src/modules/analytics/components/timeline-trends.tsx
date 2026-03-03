"use client";
import { Box, Flex, Text, Wrapper } from "ui";
import type { TooltipProps } from "recharts";
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  Legend,
} from "recharts";
import { useMemo } from "react";
import { useTheme } from "next-themes";
import { useTimelineTrends } from "../hooks/timeline-trends";
import type { StoryCompletionPoint, ObjectiveProgressPoint } from "../types";

type ChartDataItem = {
  date: string;
  storiesCompleted: number;
  storiesCreated: number;
  objectivesCompleted: number;
  totalObjectives: number;
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
    <Box className="border-border bg-surface-elevated/80 text-foreground z-50 min-w-32 rounded-lg border px-3 py-3 text-[0.95rem] font-medium backdrop-blur">
      <Text className="mb-2 font-medium">{label}</Text>
      {payload.map((entry) => (
        <Flex align="center" gap={2} key={String(entry.dataKey)}>
          <Box
            className="h-3 w-3 rounded"
            style={{ backgroundColor: entry.color }}
          />
          <Text className="capitalize">
            {entry.name}: {entry.value}
          </Text>
        </Flex>
      ))}
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

export const TimelineTrends = () => {
  const { resolvedTheme } = useTheme();
  const { data: timelineTrends, isPending } = useTimelineTrends();
  const chartData = useMemo<ChartDataItem[]>(() => {
    if (!timelineTrends) {
      return [];
    }

    const dateMap = new Map<string, ChartDataItem & { sortDate: string }>();

    timelineTrends.storyCompletion.forEach((item: StoryCompletionPoint) => {
      dateMap.set(item.date, {
        date: formatDate(item.date),
        sortDate: item.date,
        storiesCompleted: item.completed,
        storiesCreated: item.created,
        objectivesCompleted: 0,
        totalObjectives: 0,
      });
    });

    timelineTrends.objectiveProgress.forEach((item: ObjectiveProgressPoint) => {
      const existing = dateMap.get(item.date);
      if (existing) {
        existing.objectivesCompleted = item.completedObjectives;
        existing.totalObjectives = item.totalObjectives;
        return;
      }

      dateMap.set(item.date, {
        date: formatDate(item.date),
        sortDate: item.date,
        storiesCompleted: 0,
        storiesCreated: 0,
        objectivesCompleted: item.completedObjectives,
        totalObjectives: item.totalObjectives,
      });
    });

    return Array.from(dateMap.values())
      .sort(
        (a, b) =>
          new Date(a.sortDate).getTime() - new Date(b.sortDate).getTime(),
      )
      .map((dataPoint) => ({
        date: dataPoint.date,
        storiesCompleted: dataPoint.storiesCompleted,
        storiesCreated: dataPoint.storiesCreated,
        objectivesCompleted: dataPoint.objectivesCompleted,
        totalObjectives: dataPoint.totalObjectives,
      }));
  }, [timelineTrends]);

  if (isPending) {
    return (
      <Wrapper className="my-4">
        <Box className="mb-6">
          <Text className="mb-1" fontSize="lg">
            Timeline trends
          </Text>
          <Text color="muted">
            Historical trends for stories and objectives.
          </Text>
        </Box>
        <Box className="bg-skeleton h-[380px] animate-pulse rounded" />
      </Wrapper>
    );
  }

  return (
    <Wrapper className="my-4">
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Timeline trends
        </Text>
        <Text color="muted">Historical trends for stories and objectives.</Text>
      </Box>

      <ResponsiveContainer height={380} width="100%">
        <BarChart
          data={chartData}
          margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
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
          <Tooltip content={<CustomTooltip />} />
          <Legend />
          <Bar
            dataKey="storiesCompleted"
            fill="#6366F1"
            name="Stories Completed"
            radius={[4, 4, 0, 0]}
          />
          <Bar
            dataKey="storiesCreated"
            fill="#22c55e"
            name="Stories Created"
            radius={[4, 4, 0, 0]}
          />
          <Bar
            dataKey="objectivesCompleted"
            fill="#eab308"
            name="Objectives Completed"
            radius={[4, 4, 0, 0]}
          />
        </BarChart>
      </ResponsiveContainer>
    </Wrapper>
  );
};
