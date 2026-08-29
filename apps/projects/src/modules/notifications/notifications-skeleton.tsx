import { Box, Flex, Skeleton } from "ui";

export const NotificationsSkeleton = () => {
  return (
    <Box className="border-border flex h-full min-h-0 flex-col overflow-hidden border-r-[0.5px] pb-6">
      <Flex
        align="center"
        className="border-border h-16 shrink-0 border-b-[0.5px] px-4"
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
      <Box className="min-h-0 flex-1 overflow-y-auto">
        {Array.from({ length: 8 }).map((_, index) => (
          <Flex
            className="border-border/70 d/60 border-b px-4 py-3"
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
