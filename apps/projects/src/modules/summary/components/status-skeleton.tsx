/* eslint-disable react/no-array-index-key -- ok for skeletons */
"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

export const StatusSkeleton = () => {
  return (
    <Wrapper>
      <Box className="mb-6">
        <Skeleton className="mb-2 h-6 w-36 rounded" />
        <Skeleton className="h-4 w-56 rounded" />
      </Box>

      <Box>
        {/* Donut chart skeleton */}
        <Box className="relative h-[160px] w-full">
          <Skeleton className="absolute left-1/2 top-1/2 size-[160px] -translate-x-1/2 -translate-y-1/2 transform rounded-full" />
          <Skeleton className="absolute left-1/2 top-1/2 size-[100px] -translate-x-1/2 -translate-y-1/2 transform rounded-full bg-white dark:bg-dark-200" />
          {/* Center text skeleton */}
          <Flex className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 transform flex-col items-center">
            <Skeleton className="h-8 w-10 rounded" />
            <Skeleton className="mt-1 h-4 w-16 rounded" />
          </Flex>
        </Box>

        {/* Legend skeleton */}
        <Flex className="line-clamp-2 h-[60px] pt-3" gap={3} wrap>
          {Array.from({ length: 5 }).map((_, index) => (
            <Flex align="center" gap={1} key={index}>
              <Skeleton className="size-4 rounded" />
              <Skeleton className="h-4 w-12 rounded" />
            </Flex>
          ))}
        </Flex>
      </Box>
    </Wrapper>
  );
};
