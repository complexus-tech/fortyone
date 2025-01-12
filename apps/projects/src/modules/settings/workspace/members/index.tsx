"use client";

import { Box, Flex, Text, Button, Avatar, Select } from "ui";
import { TeamIcon } from "icons";
import { SectionHeader } from "../../components";

export const WorkspaceMembersSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Members
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button color="tertiary" leftIcon={<TeamIcon />} variant="outline">
              Invite Members
            </Button>
          }
          description="Manage members of your workspace and their roles."
          title="Workspace members"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            {/* Example member row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Avatar
                  className="h-8 w-8"
                  name="John Doe"
                  src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                />
                <Box>
                  <Text className="font-medium">Joseph Mukorivo</Text>
                  <Text color="muted">josemukorivo</Text>
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
                    <Select.Option value="guest">Guest</Select.Option>
                  </Select.Content>
                </Select>
                <Button
                  className="bg-primary px-3 dark:bg-primary"
                  size="sm"
                  variant="naked"
                >
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
                  <Text color="muted">jane</Text>
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
                    <Select.Option value="guest">Guest</Select.Option>
                  </Select.Content>
                </Select>
                <Button
                  className="bg-primary px-3 dark:bg-primary"
                  size="sm"
                  variant="naked"
                >
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
