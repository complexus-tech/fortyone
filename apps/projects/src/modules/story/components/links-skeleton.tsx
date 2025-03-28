"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

export const LinksSkeleton = () => {
  return (
    <Wrapper className="mb-4 mt-2">
      <Flex align="center" className="mb-4" justify="between">
        <Skeleton className="h-6 w-24 rounded" />
        <Skeleton className="h-8 w-8 rounded-full" />
      </Flex>
      <Box className="pl-2">
        <Flex align="center" className="mb-4" gap={3}>
          <Skeleton className="h-5 w-5 rounded" />
          <Skeleton className="h-5 w-48 rounded" />
        </Flex>
      </Box>
    </Wrapper>
  );
};
