/* eslint-disable react/no-array-index-key -- ok for skeleton */
"use client";
import { Box, Container, Divider, Flex, Skeleton, Wrapper } from "ui";
import { KeyResultsSkeleton } from "./key-results-skeleton";

export const OverviewSkeleton = () => {
  return (
    <Container className="h-[calc(100vh-7.7rem)] overflow-y-auto pt-6">
      <Box>
        <Flex align="center" gap={6} justify="between">
          {/* Title skeleton */}
          <Skeleton className="h-7 w-7/12 rounded md:h-10 md:w-3/4" />
          <Skeleton className="h-7 w-20 rounded md:h-8" />
        </Flex>

        {/* Description skeleton */}
        <Box className="mt-12">
          <Skeleton className="mb-3 h-4 w-full rounded" />
          <Skeleton className="mb-3 h-4 w-11/12 rounded" />
          <Skeleton className="h-4 w-5/6 rounded" />
        </Box>
      </Box>

      {/* Properties section */}
      <Flex align="center" className="mt-10" gap={2} wrap>
        <Skeleton className="h-6 w-24 rounded" />
        <Skeleton className="h-6 w-24 rounded" />
        <Skeleton className="h-6 w-24 rounded" />
        <Skeleton className="h-6 w-24 rounded" />
      </Flex>

      <Divider className="my-8" />

      {/* Activity section */}
      <Box className="mb-6">
        <Flex align="center" className="mb-4" justify="between">
          <Flex align="center" gap={2} justify="between">
            <Skeleton className="h-6 w-20 rounded" />
            <Skeleton className="h-6 w-20 rounded" />
          </Flex>
          <Skeleton className="h-6 w-24 rounded" />
        </Flex>

        <Box className="mt-6 grid w-full grid-cols-2 gap-6 md:grid-cols-3">
          {Array.from({ length: 3 }).map((_, index) => (
            <Wrapper key={index}>
              <Skeleton className="h-5 w-32 rounded" />
              <Skeleton className="mt-2 h-8 w-8 rounded md:mt-6 md:h-10" />
            </Wrapper>
          ))}
        </Box>
      </Box>

      <Flex align="center" className="mb-4" justify="between">
        <Skeleton className="h-6 w-32 rounded" />
        <Skeleton className="h-6 w-24 rounded md:h-8" />
      </Flex>
      <Divider className="mb-6" />
      {/* Key Results section */}
      <Box className="mb-6">
        <KeyResultsSkeleton />
      </Box>
    </Container>
  );
};
