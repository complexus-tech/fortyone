"use client";
import { Box, Flex, Text, Wrapper } from "ui";
import type { TooltipProps } from "recharts";
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  LineChart,
  Line,
  Legend,
} from "recharts";
import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { useTimelineTrends } from "../hooks/timeline-trends";
import type { StoryCompletionPoint, ObjectiveProgressPoint } from "../types";
import { TimelineTrendsSkeleton } from "./timeline-trends-skeleton";

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
    <Box className="z-50 min-w-32 rounded-[0.6rem] border border-gray-100 bg-white/80 px-3 py-3 text-[0.95rem] font-medium text-gray backdrop-blur dark:border-dark-50 dark:bg-dark-200 dark:text-gray-200">
      <Text className="mb-2 font-medium">{label}</Text>
      {payload.map((entry, index) => (
        <Flex align="center" gap={2} key={index}>
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
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);

  useEffect(() => {
    if (
      timelineTrends?.storyCompletion.length &&
      timelineTrends.objectiveProgress.length
    ) {
      // Combine the story and objective data by date
      const dateMap = new Map<string, ChartDataItem>();

      // Add story completion data
      timelineTrends.storyCompletion.forEach((item: StoryCompletionPoint) => {
        const formattedDate = formatDate(item.date);
        dateMap.set(formattedDate, {
          date: formattedDate,
          storiesCompleted: item.completed,
          storiesCreated: item.created,
          objectivesCompleted: 0,
          totalObjectives: 0,
        });
      });

      // Add objective progress data
      timelineTrends.objectiveProgress.forEach(
        (item: ObjectiveProgressPoint) => {
          const formattedDate = formatDate(item.date);
          const existing = dateMap.get(formattedDate);
          if (existing) {
            existing.objectivesCompleted = item.completedObjectives;
            existing.totalObjectives = item.totalObjectives;
          } else {
            dateMap.set(formattedDate, {
              date: formattedDate,
              storiesCompleted: 0,
              storiesCreated: 0,
              objectivesCompleted: item.completedObjectives,
              totalObjectives: item.totalObjectives,
            });
          }
        },
      );

      setChartData(
        Array.from(dateMap.values()).sort(
          (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime(),
        ),
      );
    }
  }, [timelineTrends]);

  if (isPending) {
    return <TimelineTrendsSkeleton />;
  }

  return (
    <Wrapper className="my-4">
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Timeline trends
        </Text>
        <Text color="muted">Historical trends for stories and objectives.</Text>
      </Box>

      <ResponsiveContainer height={280} width="100%">
        <LineChart
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
          <Line
            activeDot={{ r: 4 }}
            dataKey="storiesCompleted"
            dot={{ r: 3 }}
            name="Stories Completed"
            stroke="#6366F1"
            strokeWidth={2}
            type="monotone"
          />
          <Line
            activeDot={{ r: 4 }}
            dataKey="storiesCreated"
            dot={{ r: 3 }}
            name="Stories Created"
            stroke="#22c55e"
            strokeWidth={2}
            type="monotone"
          />
          <Line
            activeDot={{ r: 4 }}
            dataKey="objectivesCompleted"
            dot={{ r: 3 }}
            name="Objectives Completed"
            stroke="#eab308"
            strokeWidth={2}
            type="monotone"
          />
        </LineChart>
      </ResponsiveContainer>
    </Wrapper>
  );
};
