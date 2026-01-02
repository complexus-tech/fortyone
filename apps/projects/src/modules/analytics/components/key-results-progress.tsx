"use client";
import { Box, Flex, Text, Wrapper } from "ui";
import { useState, useEffect } from "react";
import { useObjectiveProgress } from "../hooks/objective-progress";
import type { KeyResultProgressItem } from "../types";
import { KeyResultsProgressSkeleton } from "./key-results-progress-skeleton";

export const KeyResultsProgress = () => {
  const { data: objectiveProgress, isPending } = useObjectiveProgress();
  const [progressData, setProgressData] = useState<KeyResultProgressItem[]>([]);

  useEffect(() => {
    if (objectiveProgress?.keyResultsProgress.length) {
      setProgressData(objectiveProgress.keyResultsProgress.slice(0, 6));
    }
  }, [objectiveProgress]);

  if (isPending) {
    return <KeyResultsProgressSkeleton />;
  }

  const getProgressColor = (progress: number) => {
    if (progress >= 80) return "bg-green-500";
    if (progress >= 60) return "bg-blue-500";
    if (progress >= 40) return "bg-yellow-500";
    return "bg-red-500";
  };

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Key results progress
        </Text>
        <Text color="muted">
          Progress tracking for key results by objective.
        </Text>
      </Box>

      <Box className="space-y-4">
        {progressData.map((item) => (
          <Box key={item.objectiveId}>
            <Flex align="center" className="mb-2" justify="between">
              <Text className="truncate pr-2 font-medium">
                {item.objectiveName}
              </Text>
              <Text color="muted" fontSize="sm">
                {Math.round(item.avgProgress)}%
              </Text>
            </Flex>
            <Box className="mb-1">
              <Box className="relative h-3 overflow-hidden rounded-full bg-surface-muted">
                <Box
                  className={`h-full rounded-full transition-all duration-300 ${getProgressColor(item.avgProgress)}`}
                  style={{
                    width: `${Math.min(item.avgProgress, 100)}%`,
                  }}
                />
              </Box>
            </Box>
            <Flex align="center" justify="between">
              <Text color="muted" fontSize="sm">
                {item.completed} of {item.total} completed
              </Text>
              <Text color="muted" fontSize="sm">
                {item.total - item.completed} remaining
              </Text>
            </Flex>
          </Box>
        ))}
      </Box>

      {progressData.length === 0 && (
        <Box className="text-text-muted flex h-[180px] items-center justify-center">
          <Text>No key results data available</Text>
        </Box>
      )}
    </Wrapper>
  );
};
