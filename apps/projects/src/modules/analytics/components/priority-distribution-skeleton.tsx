"use client";
import { Box, Wrapper } from "ui";

export const PriorityDistributionSkeleton = () => {
  return (
    <Wrapper>
      <Box className="mb-6">
        <Box className="mb-1 h-5 w-32 animate-pulse rounded bg-gray-200 dark:bg-dark-50" />
        <Box className="h-4 w-48 animate-pulse rounded bg-gray-200 dark:bg-dark-50" />
      </Box>
      <Box className="h-[220px] w-full animate-pulse rounded bg-gray-200 dark:bg-dark-50" />
    </Wrapper>
  );
};
