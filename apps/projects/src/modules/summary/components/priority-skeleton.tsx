"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

const PRIORITY_SKELETON_BAR_HEIGHTS = [132, 88, 158, 104];

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
                    height: `${PRIORITY_SKELETON_BAR_HEIGHTS[index]}px`,
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
