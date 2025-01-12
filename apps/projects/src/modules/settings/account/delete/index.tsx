"use client";

import { Box, Text, Button, Divider } from "ui";
import { SectionHeader } from "../../components";

export const DeleteAccountSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Delete Account
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Once you delete your account, there is no going back. Please be certain."
          title="Delete your account"
        />
        <Box className="p-6">
          <Text color="muted">Deleting your account will:</Text>
          <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
            <li>Delete your profile and personal information</li>
            <li>Remove you from all teams and workspaces</li>
          </ul>
          <Divider className="my-4" />
          <Button>Delete Account</Button>
        </Box>
      </Box>
    </Box>
  );
};
