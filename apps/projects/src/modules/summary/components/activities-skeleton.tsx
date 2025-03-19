/* eslint-disable react/no-array-index-key -- ok for skeletons */
"use client";
import { Flex, Skeleton, Wrapper, Text } from "ui";

const ActivityItemSkeleton = () => (
  <Flex align="center" className="mb-1 mt-4" gap={3}>
    <Skeleton className="size-8 shrink-0 rounded-full" />
    <Skeleton className="h-4 w-2/5 rounded" />
    <Skeleton className="h-4 w-16 rounded" />
    <Skeleton className="h-4 w-24 rounded" />
  </Flex>
);

export const ActivitiesSkeleton = () => {
  return (
    <Wrapper>
      <Flex align="center" className="mb-5" justify="between">
        <Text fontSize="lg">Recent activities</Text>
      </Flex>
      {Array.from({ length: 9 }).map((_, index) => (
        <ActivityItemSkeleton key={index} />
      ))}
    </Wrapper>
  );
};
