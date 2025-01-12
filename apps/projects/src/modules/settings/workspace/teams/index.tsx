"use client";

import { Box, Flex, Text, Button, Input, Avatar, Badge, Menu } from "ui";
import { SearchIcon, MoreHorizontalIcon } from "icons";
import { SectionHeader } from "../../components/section-header";

export const TeamsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Teams
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={<Button color="primary">Create Team</Button>}
          description="Create and manage teams in your workspace."
          title="Team Management"
        />

        <Box className="border-b border-gray-100 p-4 dark:border-dark-100">
          <Box className="relative max-w-md">
            <SearchIcon className="text-gray-400 absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" />
            <Input
              className="pl-9"
              placeholder="Search teams..."
              type="search"
            />
          </Box>
        </Box>

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            {/* Example team row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Box className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  🚀
                </Box>
                <Box>
                  <Text className="font-medium">Engineering</Text>
                  <Text color="muted">8 members</Text>
                </Box>
              </Flex>
              <Flex align="center" gap={4}>
                <Flex className="-space-x-2">
                  <Avatar
                    className="h-6 w-6 ring-2 ring-white dark:ring-dark-100"
                    name="John Doe"
                  />
                  <Avatar
                    className="h-6 w-6 ring-2 ring-white dark:ring-dark-100"
                    name="Jane Smith"
                  />
                  <Badge className="ml-2" color="tertiary">
                    +6
                  </Badge>
                </Flex>
                <Menu>
                  <Menu.Button>
                    <Button
                      aria-label="More options"
                      color="tertiary"
                      variant="naked"
                    >
                      <MoreHorizontalIcon className="h-4 w-4" />
                    </Button>
                  </Menu.Button>
                  <Menu.Items>
                    <Menu.Item>Edit team</Menu.Item>
                    <Menu.Item>Manage members</Menu.Item>
                    <Menu.Item>Team settings</Menu.Item>
                    <Menu.Separator />
                    <Menu.Item>Leave team</Menu.Item>
                    <Menu.Separator />
                    <Menu.Item className="text-red-600 dark:text-red-400">
                      Delete team
                    </Menu.Item>
                  </Menu.Items>
                </Menu>
              </Flex>
            </Flex>

            {/* Example team row */}
            <Flex align="center" justify="between">
              <Flex align="center" gap={3}>
                <Box className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  💼
                </Box>
                <Box>
                  <Text className="font-medium">Product</Text>
                  <Text color="muted">5 members</Text>
                </Box>
              </Flex>
              <Flex align="center" gap={4}>
                <Flex className="-space-x-2">
                  <Avatar
                    className="h-6 w-6 ring-2 ring-white dark:ring-dark-100"
                    name="Alice Johnson"
                  />
                  <Avatar
                    className="h-6 w-6 ring-2 ring-white dark:ring-dark-100"
                    name="Charlie Brown"
                  />
                  <Badge className="ml-2" color="tertiary">
                    +3
                  </Badge>
                </Flex>
                <Menu>
                  <Menu.Button>
                    <Button
                      aria-label="More options"
                      color="tertiary"
                      variant="naked"
                    >
                      <MoreHorizontalIcon className="h-4 w-4" />
                    </Button>
                  </Menu.Button>
                  <Menu.Items>
                    <Menu.Item>Edit team</Menu.Item>
                    <Menu.Item>Manage members</Menu.Item>
                    <Menu.Item>Team settings</Menu.Item>
                    <Menu.Separator />
                    <Menu.Item>Leave team</Menu.Item>
                    <Menu.Separator />
                    <Menu.Item className="text-red-600 dark:text-red-400">
                      Delete team
                    </Menu.Item>
                  </Menu.Items>
                </Menu>
              </Flex>
            </Flex>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Configure team roles and access permissions."
          title="Team Permissions"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Team Roles</Text>
              <Text className="mt-1" color="muted">
                Define roles and responsibilities within teams.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Roles
              </Button>
            </Box>

            <Box>
              <Text className="font-medium">Access Control</Text>
              <Text className="mt-1" color="muted">
                Set project access levels for different teams.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Access
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="View and analyze team performance metrics."
          title="Team Analytics"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Performance Metrics</Text>
              <Text className="mt-1" color="muted">
                Track team productivity and project progress.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                View Analytics
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
