/* eslint-disable react/no-array-index-key -- ok for skeleton */
"use client";
import { Box, Flex, Skeleton } from "ui";
import { BodyContainer } from "@/components/shared/body";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { TableHeader } from "./heading";

export const ObjectivesSkeleton = ({ isInTeam }: { isInTeam?: boolean }) => {
  return (
    <BodyContainer className="h-[calc(100vh-3.7rem)]">
      <TableHeader isInTeam={isInTeam} />
      {Array.from({ length: 8 }).map((_, index) => (
        <RowWrapper className="px-5 py-3 md:px-12" key={index}>
          <Box className="flex w-[300px] shrink-0 items-center gap-2">
            <Flex
              align="center"
              className="size-8 shrink-0 rounded-lg bg-gray-100/50 dark:bg-dark-200"
              justify="center"
            >
              <Skeleton className="h-4 w-4" />
            </Flex>
            <Skeleton className="h-6 w-56" />
          </Box>
          <Flex align="center" gap={4}>
            {!isInTeam && (
              <Box className="flex w-[45px] shrink-0 items-center gap-1.5">
                <Skeleton className="h-3 w-3 rounded-full" />
                <Skeleton className="h-5 w-8" />
              </Box>
            )}
            <Box className="flex w-[40px] shrink-0 items-center">
              <Skeleton className="h-6 w-6 rounded-full" />
            </Box>
            <Box className="flex w-[60px] shrink-0 items-center gap-1.5 pl-0.5">
              <Skeleton className="h-4 w-4 rounded-full" />
              <Skeleton className="h-5 w-10" />
            </Box>
            <Box className="w-[120px] shrink-0">
              <Skeleton className="h-5 w-24" />
            </Box>
            <Box className="w-[100px] shrink-0">
              <Skeleton className="h-5 w-20" />
            </Box>
            <Box className="w-[100px] shrink-0">
              <Skeleton className="h-5 w-24" />
            </Box>
            <Box className="w-[120px] shrink-0">
              <Skeleton className="h-5 w-20" />
            </Box>
          </Flex>
        </RowWrapper>
      ))}
    </BodyContainer>
  );
};
