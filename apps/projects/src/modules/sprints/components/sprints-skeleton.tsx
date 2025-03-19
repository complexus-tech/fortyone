/* eslint-disable react/no-array-index-key -- ok for skeleton */
"use client";
import { Flex, Box, Skeleton } from "ui";
import { BodyContainer } from "@/components/shared/body";
import { RowWrapper } from "@/components/ui/row-wrapper";

export const SprintsSkeleton = () => {
  return (
    <BodyContainer>
      {Array.from({ length: 10 }).map((_, index) => (
        <RowWrapper key={index}>
          {/* Left side - Sprint name and dates */}
          <Flex className="flex-1 items-center gap-4">
            <Flex
              align="center"
              className="size-10 rounded-lg bg-gray-100/50 dark:bg-dark-200"
              justify="center"
            >
              <Skeleton className="h-5 w-5" />
            </Flex>
            <Box className="space-y-2">
              <Skeleton className="h-4 w-40" />
              <Flex className="items-center gap-1.5">
                <Skeleton className="h-4 w-4" /> {/* Calendar icon */}
                <Skeleton className="h-4 w-16" /> {/* Start date */}
                <Skeleton className="h-3 w-3" /> {/* Arrow icon */}
                <Skeleton className="h-4 w-16" /> {/* End date */}
              </Flex>
            </Box>
          </Flex>

          {/* Right side - Status, progress and metrics */}
          <Flex className="items-center" gap={4}>
            {/* Status badge */}
            <Skeleton className="h-5 w-24 rounded-full" />

            {/* Progress bar */}
            <Flex align="center" className="w-36" gap={3}>
              <Skeleton className="h-2 flex-1 rounded-full" />
              <Skeleton className="h-5 w-10" />
            </Flex>

            {/* Story counts */}
            <Flex className="min-w-[300px]" gap={4}>
              {/* Done */}
              <Flex align="center" className="min-w-[80px] gap-1.5">
                <Skeleton className="h-4 w-4 rounded-full" />
                <Skeleton className="h-5 w-16" />
              </Flex>

              {/* Active */}
              <Flex align="center" className="min-w-[80px] gap-1.5">
                <Skeleton className="h-4 w-4 rounded-full" />
                <Skeleton className="h-5 w-16" />
              </Flex>

              {/* Todo */}
              <Flex align="center" className="min-w-[80px] gap-1.5">
                <Skeleton className="h-4 w-4 rounded-full" />
                <Skeleton className="h-5 w-16" />
              </Flex>

              {/* Backlog */}
              <Flex align="center" className="min-w-[80px] gap-1.5">
                <Skeleton className="h-4 w-4 rounded-full" />
                <Skeleton className="h-5 w-16" />
              </Flex>
            </Flex>
          </Flex>
        </RowWrapper>
      ))}
    </BodyContainer>
  );
};
