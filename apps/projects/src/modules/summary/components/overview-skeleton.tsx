"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

const SkeletonCard = () => (
  <Wrapper>
    <Flex justify="between">
      <Skeleton className="mb-5 h-7 w-6 rounded" />
      <Skeleton className="h-5 w-5 rounded" />
    </Flex>
    <Skeleton className="h-4 w-24 rounded" />
  </Wrapper>
);

export const OverviewSkeleton = () => {
  return (
    <Box className="mb-4 mt-3 grid grid-cols-2 gap-4 md:grid-cols-5">
      {Array.from({ length: 5 }).map((_, index) => (
        <SkeletonCard key={index} />
      ))}
    </Box>
  );
};
