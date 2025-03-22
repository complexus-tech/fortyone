/* eslint-disable react/no-array-index-key -- ok for skeleton */
"use client";
import { Box, Container, Divider, Flex, Skeleton, Wrapper } from "ui";
import { BodyContainer } from "@/components/shared";
import { ActivitiesSkeleton } from "./activities-skeleton";
import { LinksSkeleton } from "./links-skeleton";
import { AttachmentsSkeleton } from "./attachments-skeleton";

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

        <LinksSkeleton />

        <AttachmentsSkeleton />

        <Divider className="my-6" />

        <ActivitiesSkeleton />
      </Container>
    </BodyContainer>
  );
};
