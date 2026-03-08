"use client";
import { Box, Wrapper } from "ui";

export const KeyResultsProgressSkeleton = () => {
  return (
    <Wrapper>
      <Box className="mb-6">
        <Box className="bg-skeleton mb-1 h-5 w-32 animate-pulse rounded" />
        <Box className="bg-skeleton h-4 w-48 animate-pulse rounded" />
      </Box>
      <Box className="space-y-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <Box key={index}>
            <Box className="bg-skeleton mb-2 h-4 w-full animate-pulse rounded" />
            <Box className="bg-skeleton mb-1 h-3 w-full animate-pulse rounded-full" />
            <Box className="bg-skeleton h-3 w-32 animate-pulse rounded" />
          </Box>
        ))}
      </Box>
    </Wrapper>
  );
};
