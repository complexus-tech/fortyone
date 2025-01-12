"use client";

import { Box, Flex, Text, Button, Input } from "ui";

export const WorkspaceDeleteSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Delete Workspace
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
          <Text as="h3" className="font-medium">
            Delete your workspace
          </Text>
          <Text className="mt-1" color="muted">
            Once you delete your workspace, there is no going back. Please be
            certain.
          </Text>
        </Box>

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">Deleting your workspace will:</Text>
              <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
                <li>Delete all teams and their data</li>
                <li>Delete all stories, comments, and activities</li>
                <li>Remove all members from the workspace</li>
                <li>Delete all workspace settings and configurations</li>
              </ul>
            </Box>

            <Box>
              <Input
                helpText="Please type your workspace name to confirm deletion"
                label="Confirm workspace name"
                name="workspace_name"
                placeholder="Enter workspace name to confirm"
              />
            </Box>

            <Box>
              <Button color="danger" variant="outline">
                Delete Workspace
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
