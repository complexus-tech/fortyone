"use client";

import { Box, Flex, Text, Button } from "ui";

export const WorkspaceWorkflowsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Workflows
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 p-6 dark:border-dark-100">
          <Flex align="center" justify="between">
            <Box>
              <Text as="h3" className="font-medium">
                Story workflows
              </Text>
              <Text className="mt-1" color="muted">
                Configure how stories move through your workflow.
              </Text>
            </Box>
            <Button color="primary" variant="outline">
              Create Workflow
            </Button>
          </Flex>
        </Box>

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            {/* Example workflow row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Box className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  ⚡️
                </Box>
                <Box>
                  <Text className="font-medium">Default Workflow</Text>
                  <Flex align="center" className="mt-1" gap={2}>
                    <Box className="h-2 w-2 rounded-full bg-gray-300" />
                    <Box className="bg-yellow-400 h-2 w-2 rounded-full" />
                    <Box className="bg-blue-400 h-2 w-2 rounded-full" />
                    <Box className="bg-green-400 h-2 w-2 rounded-full" />
                    <Text color="muted">4 statuses</Text>
                  </Flex>
                </Box>
              </Flex>
              <Flex align="center" gap={3}>
                <Button color="tertiary" variant="outline">
                  Edit
                </Button>
              </Flex>
            </Flex>

            {/* Example workflow row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Box className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  🎯
                </Box>
                <Box>
                  <Text className="font-medium">Bug Workflow</Text>
                  <Flex align="center" className="mt-1" gap={2}>
                    <Box className="h-2 w-2 rounded-full bg-gray-300" />
                    <Box className="bg-red-400 h-2 w-2 rounded-full" />
                    <Box className="bg-blue-400 h-2 w-2 rounded-full" />
                    <Box className="bg-green-400 h-2 w-2 rounded-full" />
                    <Text color="muted">4 statuses</Text>
                  </Flex>
                </Box>
              </Flex>
              <Flex align="center" gap={3}>
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
