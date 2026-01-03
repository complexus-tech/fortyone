"use client";
import { Box, Wrapper } from "ui";

export const TeamWorkloadSkeleton = () => {
  return (
    <Wrapper className="my-4">
      <Box className="mb-6">
        <Box className="mb-1 h-5 w-32 animate-pulse rounded bg-skeleton" />
        <Box className="h-4 w-48 animate-pulse rounded bg-skeleton" />
      </Box>
      <Box className="h-[280px] w-full animate-pulse rounded bg-skeleton" />
    </Wrapper>
  );
};
