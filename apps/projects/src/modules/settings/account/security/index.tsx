"use client";

import { Box, Flex, Text, Button, Switch } from "ui";
import { SectionHeader } from "../../components";

export const SecuritySettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Security Settings
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Add an extra layer of security to your account."
          title="Two-Factor Authentication"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Flex align="center" justify="between">
              <Box>
                <Text className="font-medium">Enable 2FA</Text>
                <Text color="muted">
                  Protect your account with two-factor authentication
                </Text>
              </Box>
              <Switch />
            </Flex>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Manage your password and security preferences."
          title="Password"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Button color="tertiary" variant="outline">
                Change Password
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Manage your active sessions and devices."
          title="Active Sessions"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No active sessions found.</Text>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
