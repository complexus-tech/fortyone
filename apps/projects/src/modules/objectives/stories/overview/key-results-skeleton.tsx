"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

export const KeyResultsSkeleton = () => {
  return (
    <Box>
      {Array.from({ length: 3 }).map((_, index) => (
        <Wrapper
          className="mb-4 flex items-center justify-between py-6"
          key={index}
        >
          <Flex align="center" gap={2}>
            <Skeleton className="size-8 rounded" />
            <Skeleton className="h-5 w-16 rounded" />
          </Flex>
          <Flex align="center" gap={2}>
            <Skeleton className="h-5 w-24 rounded" />
            <Skeleton className="h-5 w-24 rounded" />
            <Skeleton className="size-8 rounded-full" />
          </Flex>
        </Wrapper>
      ))}
    </Box>
  );
};
