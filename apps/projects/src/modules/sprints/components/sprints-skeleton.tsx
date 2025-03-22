/* eslint-disable react/no-array-index-key -- ok for skeleton */
"use client";
import { Flex, Box, Skeleton } from "ui";
import { BodyContainer } from "@/components/shared/body";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { SprintsHeader } from "./header";

export const SprintsSkeleton = () => {
  return (
    <>
      <SprintsHeader />
      <BodyContainer>
        {Array.from({ length: 10 }).map((_, index) => (
          <RowWrapper key={index}>
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
                  <Skeleton className="h-4 w-4" />
                  <Skeleton className="h-4 w-16" />
                  <Skeleton className="h-3 w-3" />
                  <Skeleton className="h-4 w-16" />
                </Flex>
              </Box>
            </Flex>
            <Flex className="items-center" gap={4}>
              <Skeleton className="h-5 w-24 rounded-full" />
              <Flex align="center" className="w-36" gap={3}>
                <Skeleton className="h-2 flex-1 rounded-full" />
                <Skeleton className="h-5 w-10" />
              </Flex>
              <Flex className="min-w-[300px]" gap={4}>
                <Flex align="center" className="min-w-[80px] gap-1.5">
                  <Skeleton className="h-4 w-4 rounded-full" />
                  <Skeleton className="h-5 w-16" />
                </Flex>
                <Flex align="center" className="min-w-[80px] gap-1.5">
                  <Skeleton className="h-4 w-4 rounded-full" />
                  <Skeleton className="h-5 w-16" />
                </Flex>
                <Flex align="center" className="min-w-[80px] gap-1.5">
                  <Skeleton className="h-4 w-4 rounded-full" />
                  <Skeleton className="h-5 w-16" />
                </Flex>
                <Flex align="center" className="min-w-[80px] gap-1.5">
                  <Skeleton className="h-4 w-4 rounded-full" />
                  <Skeleton className="h-5 w-16" />
                </Flex>
              </Flex>
            </Flex>
          </RowWrapper>
        ))}
      </BodyContainer>
    </>
  );
};
