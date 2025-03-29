/* eslint-disable react/no-array-index-key -- ok for skeletons */
"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

export const PrioritySkeleton = () => {
  return (
    <Wrapper>
      <Box className="mb-6">
        <Skeleton className="mb-2 h-6 w-40 rounded" />
        <Skeleton className="h-4 w-64 rounded" />
      </Box>
      <Box className="h-[220px]">
        {/* Chart bars skeleton */}
        <Flex className="h-full flex-col justify-end">
          <Flex className="mt-auto h-full items-end gap-8 px-5">
            {Array.from({ length: 4 }).map((_, index) => (
              <Flex
                className="flex-1 flex-col items-center justify-end gap-2"
                key={index}
              >
                <Skeleton
                  className="w-full rounded-t"
                  style={{
                    height: `${Math.max(30, Math.random() * 150)}px`,
                  }}
                />
                <Skeleton className="h-4 w-full rounded" />
              </Flex>
            ))}
          </Flex>
        </Flex>
      </Box>
    </Wrapper>
  );
};
