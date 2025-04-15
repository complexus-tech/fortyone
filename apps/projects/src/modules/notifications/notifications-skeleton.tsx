/* eslint-disable react/no-array-index-key -- ok for skeleton */
import { Box, Flex, Skeleton } from "ui";

export const NotificationsSkeleton = () => {
  return (
    <Box className="h-dvh border-r-[0.5px] border-gray-200/60 pb-6 dark:border-dark-50">
      <Flex
        align="center"
        className="h-16 border-b-[0.5px] border-gray-200/60 px-4 dark:border-dark-50"
        justify="between"
      >
        <Flex align="center" gap={2}>
          <Skeleton className="size-5 rounded-full" />
          <Skeleton className="h-5 w-24" />
        </Flex>
        <Flex align="center" gap={2}>
          <Skeleton className="size-8 rounded-md" />
          <Skeleton className="size-8 rounded-md" />
        </Flex>
      </Flex>
      <Box className="h-[calc(100dvh-4rem)] overflow-y-auto">
        {Array.from({ length: 8 }).map((_, index) => (
          <Flex
            className="border-b border-gray-100/70 px-4 py-3 dark:border-dark-50/60"
            direction="column"
            gap={2}
            key={index}
          >
            <Flex align="center" justify="between">
              <Skeleton className="h-4 w-3/5" />
              <Skeleton className="h-4 w-10" />
            </Flex>

            <Flex align="center" justify="between">
              <Flex align="center" gap={2}>
                <Skeleton className="size-8 rounded-full" />
                <Skeleton className="h-4 w-40" />
              </Flex>
              <Skeleton className="size-4" />
            </Flex>
          </Flex>
        ))}
      </Box>
    </Box>
  );
};
