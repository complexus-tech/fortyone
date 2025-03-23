/* eslint-disable react/no-array-index-key -- ok for skeleton */
"use client";
import { Box, Container, Flex, Skeleton } from "ui";

export const SidebarSkeleton = ({ className }: { className?: string }) => {
  return (
    <Box className={className}>
      <Container className="md:px-6">
        <Flex align="center" className="mb-6 mt-6" gap={2}>
          <Skeleton className="h-5 w-24 rounded" />
          <Skeleton className="h-8 w-8 rounded-full" />
        </Flex>

        <Box className="mt-10">
          {Array.from({ length: 6 }).map((_, index) => (
            <Flex
              align="center"
              className="mb-6 px-1 py-2"
              gap={2}
              justify="between"
              key={index}
            >
              <Flex align="center" gap={2}>
                <Skeleton className="h-5 w-5 rounded-full" />
                <Skeleton className="h-5 w-32 rounded" />
              </Flex>
              <Skeleton className="h-5 w-8 rounded" />
            </Flex>
          ))}
        </Box>
      </Container>
    </Box>
  );
};
