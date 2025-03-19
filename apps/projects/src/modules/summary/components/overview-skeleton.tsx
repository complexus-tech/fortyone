/* eslint-disable react/no-array-index-key -- ok for skeletons */
"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

const SkeletonCard = () => (
  <Wrapper className="px-5">
    <Flex justify="between">
      <Skeleton className="mb-5 h-7 w-6 rounded" />
      <Skeleton className="h-5 w-5 rounded" />
    </Flex>
    <Skeleton className="h-4 w-24 rounded" />
  </Wrapper>
);

export const OverviewSkeleton = () => {
  return (
    <Box className="mb-4 mt-3 grid grid-cols-5 gap-4">
      {Array.from({ length: 5 }).map((_, index) => (
        <SkeletonCard key={index} />
      ))}
    </Box>
  );
};
