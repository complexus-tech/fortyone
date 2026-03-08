"use client";
import { Box, Wrapper } from "ui";

export const BurndownChartSkeleton = () => {
  return (
    <Wrapper>
      <Box className="mb-6">
        <Box className="bg-skeleton mb-1 h-5 w-32 animate-pulse rounded" />
        <Box className="bg-skeleton h-4 w-48 animate-pulse rounded" />
      </Box>
      <Box className="bg-skeleton h-[220px] w-full animate-pulse rounded" />
    </Wrapper>
  );
};
