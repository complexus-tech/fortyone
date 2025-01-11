"use client";

import { Box, Flex, Text, Button, Avatar, Select } from "ui";

export const WorkspaceMembersSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Members
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 p-6 dark:border-dark-100">
          <Flex align="center" justify="between">
            <Box>
              <Text as="h3" className="font-medium">
                Workspace members
              </Text>
              <Text className="mt-1" color="muted">
                Manage members of your workspace and their roles.
              </Text>
            </Box>
            <Button color="primary" variant="outline">
              Invite Members
            </Button>
          </Flex>
        </Box>

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            {/* Example member row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Avatar className="h-8 w-8" name="John Doe" />
                <Box>
                  <Text className="font-medium">John Doe</Text>
                  <Text color="muted">john@example.com</Text>
                </Box>
              </Flex>
              <Flex align="center" gap={3}>
                <Select defaultValue="member">
                  <Select.Trigger className="w-32">
                    <Select.Input />
                  </Select.Trigger>
                  <Select.Content>
                    <Select.Option value="admin">Admin</Select.Option>
                    <Select.Option value="member">Member</Select.Option>
                  </Select.Content>
                </Select>
                <Button color="danger" variant="outline">
                  Remove
                </Button>
              </Flex>
            </Flex>

            {/* Example member row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Avatar className="h-8 w-8" name="Jane Smith" />
                <Box>
                  <Text className="font-medium">Jane Smith</Text>
                  <Text color="muted">jane@example.com</Text>
                </Box>
              </Flex>
              <Flex align="center" gap={3}>
                <Select defaultValue="admin">
                  <Select.Trigger className="w-32">
                    <Select.Input />
                  </Select.Trigger>
                  <Select.Content>
                    <Select.Option value="admin">Admin</Select.Option>
                    <Select.Option value="member">Member</Select.Option>
                  </Select.Content>
                </Select>
                <Button color="danger" variant="outline">
                  Remove
                </Button>
              </Flex>
            </Flex>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
