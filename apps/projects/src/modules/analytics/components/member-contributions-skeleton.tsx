"use client";
import { Box, Flex, Wrapper } from "ui";

export const MemberContributionsSkeleton = () => {
  return (
    <Wrapper>
      <Box className="mb-6">
        <Box className="mb-1 h-5 w-32 animate-pulse rounded bg-skeleton" />
        <Box className="h-4 w-48 animate-pulse rounded bg-skeleton" />
      </Box>
      <Box className="space-y-3">
        {Array.from({ length: 6 }).map((_, index) => (
          <Flex align="center" gap={3} key={index}>
            <Box className="h-8 w-8 animate-pulse rounded-full bg-skeleton" />
            <Box className="flex-1">
              <Box className="mb-2 h-4 w-24 animate-pulse rounded bg-skeleton" />
              <Box className="h-2 w-full animate-pulse rounded-full bg-skeleton" />
            </Box>
          </Flex>
        ))}
      </Box>
    </Wrapper>
  );
};
