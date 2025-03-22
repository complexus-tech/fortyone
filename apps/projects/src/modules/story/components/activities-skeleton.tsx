"use client";
import { Box, Flex, Skeleton } from "ui";

export const ActivitiesSkeleton = () => {
  return (
    <Box>
      <Skeleton className="mb-6 h-6 w-32 rounded" />
      {Array.from({ length: 4 }).map((_, index) => (
        <Flex align="center" className="mb-5" gap={3} key={index}>
          <Skeleton className="size-8 shrink-0 rounded-full" />
          <Box className="w-full">
            <Flex align="center" gap={2}>
              <Skeleton className="h-5 w-24 rounded" />
              <Skeleton className="h-5 w-16 rounded" />
            </Flex>
            <Skeleton className="mt-1 h-4 w-2/3 rounded" />
          </Box>
        </Flex>
      ))}
    </Box>
  );
};
