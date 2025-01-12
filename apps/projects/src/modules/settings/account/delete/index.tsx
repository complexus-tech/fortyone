"use client";

import { Box, Flex, Text, Button } from "ui";

export const DeleteAccountSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Delete Account
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
          <Text as="h3" className="font-medium">
            Delete your account
          </Text>
          <Text className="mt-1" color="muted">
            Once you delete your account, there is no going back. Please be
            certain.
          </Text>
        </Box>

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">Deleting your account will:</Text>
              <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
                <li>Delete your profile and personal information</li>
                <li>Remove you from all teams and workspaces</li>
                <li>Delete all your stories, comments, and activities</li>
              </ul>
            </Box>

            <Box>
              <Button color="danger" variant="outline">
                Delete Account
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
