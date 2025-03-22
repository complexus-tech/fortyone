"use client";
import { Box, Container, Divider, Flex, Skeleton } from "ui";

const OptionSkeleton = ({ label }: { label: string }) => (
  <Flex className="mb-4" justify="between">
    <Box className="text-gray-500 dark:text-gray-400 font-medium">{label}</Box>
    <Flex align="center" gap={2}>
      <Skeleton className="h-6 w-24 rounded" />
    </Flex>
  </Flex>
);

export const OptionsSkeleton = () => {
  return (
    <Box className="h-screen overflow-y-auto">
      <Container className="pt-7">
        <Flex align="center" className="mb-4" justify="between">
          <Flex align="center" gap={2}>
            <Skeleton className="h-5 w-16 rounded" />
          </Flex>
          <Flex align="center" gap={2}>
            <Skeleton className="h-8 w-8 rounded-full" />
            <Skeleton className="h-8 w-8 rounded-full" />
          </Flex>
        </Flex>

        <Divider className="my-4" />

        <OptionSkeleton label="Assignee" />
        <OptionSkeleton label="Reporter" />
        <OptionSkeleton label="Status" />
        <OptionSkeleton label="Priority" />

        <OptionSkeleton label="Due date" />
        <OptionSkeleton label="Sprint" />
        <OptionSkeleton label="Objective" />

        <OptionSkeleton label="Labels" />

        <Divider className="my-4" />

        {/* Links section */}
        <Flex align="center" className="mb-4" justify="between">
          <Box className="text-gray-500 dark:text-gray-400 font-medium">
            Links
          </Box>
          <Skeleton className="h-8 w-8 rounded-full" />
        </Flex>

        <Flex align="center" className="mt-4" gap={2}>
          <Skeleton className="h-6 w-24 rounded" />
          <Skeleton className="h-6 w-48 rounded" />
        </Flex>
      </Container>
    </Box>
  );
};
