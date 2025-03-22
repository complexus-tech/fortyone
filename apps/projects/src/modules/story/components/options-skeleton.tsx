"use client";
import { Box, Container, Divider, Flex, Skeleton, Text } from "ui";

const OptionSkeleton = ({ label }: { label: string }) => (
  <Flex className="mb-6" justify="between">
    <Text color="muted" fontWeight="medium">
      {label}
    </Text>
    <Flex align="center" gap={2}>
      <Skeleton className="h-6 w-36 rounded" />
    </Flex>
  </Flex>
);

export const OptionsSkeleton = () => {
  return (
    <Box className="h-full overflow-y-auto bg-gradient-to-br from-white via-gray-50/50 to-gray-50 pb-6 dark:from-dark-200/50 dark:to-dark">
      <Container className="pt-7 md:px-6">
        <Flex align="center" className="mb-12" justify="between">
          <Flex align="center" gap={2}>
            <Skeleton className="h-5 w-16 rounded" />
          </Flex>
          <Flex align="center" gap={2}>
            <Skeleton className="h-8 w-8 rounded-full" />
            <Skeleton className="h-8 w-8 rounded-full" />
          </Flex>
        </Flex>

        <OptionSkeleton label="Assignee" />
        <OptionSkeleton label="Reporter" />
        <OptionSkeleton label="Status" />
        <OptionSkeleton label="Priority" />

        <OptionSkeleton label="Due date" />
        <OptionSkeleton label="Sprint" />
        <OptionSkeleton label="Objective" />

        <OptionSkeleton label="Labels" />

        <Divider className="my-8" />

        {/* Links section */}
        <Flex align="center" className="mb-4" justify="between">
          <Text color="muted" fontWeight="medium">
            Links
          </Text>
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
