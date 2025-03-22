/* eslint-disable react/no-array-index-key -- ok for skeleton */
"use client";
import { Box, Container, Divider, Flex, Skeleton, Wrapper } from "ui";
import { BodyContainer } from "@/components/shared";

export const MainDetailsSkeleton = () => {
  return (
    <BodyContainer className="h-screen overflow-y-auto pb-8">
      <Container className="pt-7">
        {/* Title skeleton */}
        <Skeleton className="h-10 w-3/4 rounded" />

        {/* Description skeleton */}
        <Box className="mt-12">
          <Skeleton className="mb-3 h-4 w-full rounded" />
          <Skeleton className="mb-3 h-4 w-11/12 rounded" />
          <Skeleton className="mb-3 h-4 w-5/6 rounded" />
          <Skeleton className="mb-3 h-4 w-3/4 rounded" />
        </Box>

        {/* SubStories skeleton */}
        <Wrapper className="mb-4 mt-10">
          <Flex align="center" className="mb-4" justify="between">
            <Skeleton className="h-6 w-32 rounded" />
            <Skeleton className="h-8 w-8 rounded-full" />
          </Flex>
          <Box className="pl-2">
            {Array.from({ length: 2 }).map((_, index) => (
              <Flex align="center" className="mb-4" gap={3} key={index}>
                <Skeleton className="h-5 w-5 rounded-full" />
                <Skeleton className="h-5 w-64 rounded" />
              </Flex>
            ))}
          </Box>
        </Wrapper>

        {/* Links skeleton */}
        <Wrapper className="mb-4">
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

        {/* Attachments skeleton */}
        <Wrapper className="mb-4 mt-2.5 border-t border-gray-100/60 pt-2.5 dark:border-dark-100/80">
          <Flex align="center" className="mb-4" justify="between">
            <Skeleton className="h-6 w-32 rounded" />
            <Skeleton className="h-8 w-8 rounded-full" />
          </Flex>
        </Wrapper>

        <Divider className="my-6" />

        {/* Activities skeleton */}
        <Box>
          <Skeleton className="mb-6 h-6 w-32 rounded" />
          {Array.from({ length: 4 }).map((_, index) => (
            <Flex align="center" className="mb-5" gap={3} key={index}>
              <Skeleton className="size-8 shrink-0 rounded-full" />
              <Box className="w-full">
                <Flex align="center" gap={2}>
                  <Skeleton className="h-5 w-24 rounded" />
                  <Skeleton className="h-5 w-16 rounded" />
                </Flex>
                <Skeleton className="mt-1 h-4 w-2/3 rounded" />
              </Box>
            </Flex>
          ))}
        </Box>
      </Container>
    </BodyContainer>
  );
};
