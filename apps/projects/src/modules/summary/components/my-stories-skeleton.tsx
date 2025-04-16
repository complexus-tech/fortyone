/* eslint-disable react/no-array-index-key -- ok for skeletons */
"use client";
import { Box, Flex, Skeleton, Tabs, Wrapper, Text } from "ui";
import { RowWrapper } from "@/components/ui";
import { useTerminology } from "@/hooks";

const StoryRowSkeleton = () => (
  <RowWrapper className="gap-4 px-0 last-of-type:border-b-0">
    <Flex align="center" className="relative" gap={2}>
      <Skeleton className="h-4 w-14 rounded" />
      <Skeleton className="h-4 w-4 rounded-full" />
      <Skeleton className="h-4 w-40 rounded" />
    </Flex>
    <Flex align="center" className="shrink-0" gap={3}>
      <Flex align="center" gap={2}>
        <Skeleton className="h-4 w-4 rounded-full" />
        <Skeleton className="h-4 w-16 rounded" />
      </Flex>
      <Skeleton className="h-6 w-6 rounded-full" />
    </Flex>
  </RowWrapper>
);

export const MyStoriesSkeleton = () => {
  const { getTermDisplay } = useTerminology();
  return (
    <Wrapper>
      <Flex align="center" justify="between">
        <Text className="mb-2" fontSize="lg">
          Recent {getTermDisplay("storyTerm", { variant: "plural" })}
        </Text>
        <Skeleton className="h-7 w-20" />
      </Flex>

      <Tabs defaultValue="inProgress">
        <Tabs.List className="mx-0 md:mx-0">
          <Tabs.Tab value="inProgress">In Progress</Tabs.Tab>
          <Tabs.Tab value="upcoming">Due soon</Tabs.Tab>
          <Tabs.Tab value="due">Overdue</Tabs.Tab>
        </Tabs.List>

        <Box className="mt-2.5 border-t border-gray-50 dark:border-dark-200">
          {Array.from({ length: 9 }).map((_, index) => (
            <StoryRowSkeleton key={index} />
          ))}
        </Box>
      </Tabs>
    </Wrapper>
  );
};
