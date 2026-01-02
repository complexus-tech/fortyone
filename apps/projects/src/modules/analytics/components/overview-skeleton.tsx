"use client";
import { Box, Wrapper } from "ui";

const CardSkeleton = () => (
  <Wrapper className="px-3 py-3 md:px-5 md:py-4">
    <Box className="mb-3 h-7 w-12 animate-pulse rounded bg-skeleton" />
    <Box className="h-4 w-20 animate-pulse rounded bg-skeleton" />
  </Wrapper>
);

export const OverviewSkeleton = () => {
  return (
    <Box className="mb-4 mt-3 grid grid-cols-2 gap-3 md:grid-cols-5 md:gap-4">
      {Array.from({ length: 5 }).map((_, index) => (
        <CardSkeleton key={index} />
      ))}
    </Box>
  );
};
