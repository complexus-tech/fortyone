"use client";
import { Box, Wrapper } from "ui";

export const ObjectiveHealthSkeleton = () => {
  return (
    <Wrapper>
      <Box className="mb-6">
        <Box className="bg-skeleton mb-1 h-5 w-32 animate-pulse rounded" />
        <Box className="bg-skeleton h-4 w-48 animate-pulse rounded" />
      </Box>
      <Box className="mb-4">
        <Box className="bg-skeleton mx-auto h-40 w-40 animate-pulse rounded-full" />
      </Box>
      <Box className="flex gap-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <Box className="flex items-center gap-1" key={index}>
            <Box className="bg-skeleton h-4 w-4 animate-pulse rounded" />
            <Box className="bg-skeleton h-4 w-16 animate-pulse rounded" />
          </Box>
        ))}
      </Box>
    </Wrapper>
  );
};
