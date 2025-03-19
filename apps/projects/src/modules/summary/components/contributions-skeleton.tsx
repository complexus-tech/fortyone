/* eslint-disable react/no-array-index-key -- ok for skeletons */
"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

export const ContributionsSkeleton = () => {
  return (
    <Wrapper className="mb-4 mt-5">
      <Flex align="center" className="mb-5" justify="between">
        <Skeleton className="h-6 w-32 rounded" />
        <Skeleton className="h-8 w-20 rounded" />
      </Flex>

      <Box className="mt-3">
        {/* Calendar/chart skeleton */}
        <Box className="mb-4 grid grid-cols-7 gap-1">
          {Array.from({ length: 28 }).map((_, index) => (
            <Skeleton
              className="aspect-square rounded"
              key={index}
              style={{ opacity: Math.max(0.2, Math.random()) }}
            />
          ))}
        </Box>

        {/* Legend/metrics skeleton */}
        <Flex className="mt-4" justify="between">
          {Array.from({ length: 4 }).map((_, index) => (
            <Flex align="center" gap={2} key={index}>
              <Skeleton className="h-3 w-3 rounded-full" />
              <Skeleton className="h-4 w-16 rounded" />
            </Flex>
          ))}
        </Flex>
      </Box>
    </Wrapper>
  );
};
