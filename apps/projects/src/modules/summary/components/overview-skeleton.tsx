"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

const SkeletonCard = () => (
  <Wrapper className="px-3 py-3 md:px-5 md:py-4">
    <Flex justify="between">
      <Skeleton className="mb-5 h-7 w-6 rounded" />
      <Skeleton className="h-5 w-5 rounded" />
    </Flex>
    <Skeleton className="h-4 w-24 rounded" />
  </Wrapper>
);

export const OverviewSkeleton = () => {
  return (
    <Box className="mt-3 mb-4 grid grid-cols-2 gap-3 @3xl:grid-cols-3 @4xl:grid-cols-4 @7xl:grid-cols-5 @7xl:gap-4">
      {Array.from({ length: 5 }).map((_, index) => (
        <SkeletonCard key={index} />
      ))}
    </Box>
  );
};
