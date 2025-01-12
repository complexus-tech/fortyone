"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../../components/section-header";

export const SecuritySettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Security Settings
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Configure authentication methods for your workspace."
          title="Authentication"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Single Sign-On (SSO)</Text>
              <Text className="mt-1" color="muted">
                Enable SSO to manage workspace access through your identity
                provider.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure SSO
              </Button>
            </Box>

            <Box>
              <Text className="font-medium">Two-Factor Authentication</Text>
              <Text className="mt-1" color="muted">
                Require 2FA for all workspace members.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Enable 2FA
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Manage access control and permissions for your workspace."
          title="Access Control"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">IP Allowlist</Text>
              <Text className="mt-1" color="muted">
                Restrict access to specific IP addresses or ranges.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure IP Allowlist
              </Button>
            </Box>

            <Box>
              <Text className="font-medium">Session Management</Text>
              <Text className="mt-1" color="muted">
                Set session duration and manage active sessions.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Manage Sessions
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="View security audit logs and activity history."
          title="Audit Logs"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No audit logs available.</Text>
              <Text className="mt-2" color="muted">
                Track important security events and user activities in your
                workspace.
              </Text>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
