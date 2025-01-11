"use client";

import { Box, Flex, Text, Button } from "ui";

export const WorkspaceLabelsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Labels
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
          <Flex align="center" justify="between">
            <Box>
              <Text as="h3" className="font-medium">
                Story labels
              </Text>
              <Text className="mt-1" color="muted">
                Create and manage labels to categorize stories.
              </Text>
            </Box>
            <Button color="primary" variant="outline">
              Create Label
            </Button>
          </Flex>
        </Box>

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            {/* Example label row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Box className="bg-blue-400 h-4 w-4 rounded" />
                <Box>
                  <Text className="font-medium">Feature</Text>
                  <Text color="muted">New feature or enhancement</Text>
                </Box>
              </Flex>
              <Flex align="center" gap={3}>
                <Text color="muted">24 stories</Text>
                <Button color="tertiary" variant="outline">
                  Edit
                </Button>
              </Flex>
            </Flex>

            {/* Example label row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Box className="bg-red-400 h-4 w-4 rounded" />
                <Box>
                  <Text className="font-medium">Bug</Text>
                  <Text color="muted">Something is not working</Text>
                </Box>
              </Flex>
              <Flex align="center" gap={3}>
                <Text color="muted">12 stories</Text>
                <Button color="tertiary" variant="outline">
                  Edit
                </Button>
              </Flex>
            </Flex>

            {/* Example label row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Box className="bg-yellow-400 h-4 w-4 rounded" />
                <Box>
                  <Text className="font-medium">Documentation</Text>
                  <Text color="muted">
                    Improvements or additions to documentation
                  </Text>
                </Box>
              </Flex>
              <Flex align="center" gap={3}>
                <Text color="muted">8 stories</Text>
                <Button color="tertiary" variant="outline">
                  Edit
                </Button>
              </Flex>
            </Flex>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
