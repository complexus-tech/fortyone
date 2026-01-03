"use client";
import { Box, Wrapper } from "ui";

export const SprintHealthSkeleton = () => {
  return (
    <Wrapper>
      <Box className="mb-6">
        <Box className="mb-1 h-5 w-32 animate-pulse rounded bg-skeleton" />
        <Box className="h-4 w-48 animate-pulse rounded bg-skeleton" />
      </Box>
      <Box className="mb-4">
        <Box className="mx-auto h-40 w-40 animate-pulse rounded-full bg-skeleton" />
      </Box>
      <Box className="flex gap-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <Box className="flex items-center gap-1" key={index}>
            <Box className="h-4 w-4 animate-pulse rounded bg-skeleton" />
            <Box className="h-4 w-16 animate-pulse rounded bg-skeleton" />
          </Box>
        ))}
      </Box>
    </Wrapper>
  );
};
