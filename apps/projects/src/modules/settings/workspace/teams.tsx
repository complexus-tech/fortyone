"use client";

import { Box, Flex, Text, Button, Avatar } from "ui";

export const WorkspaceTeamsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Teams
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 p-6 dark:border-dark-100">
          <Flex align="center" justify="between">
            <Box>
              <Text as="h3" className="font-medium">
                Workspace teams
              </Text>
              <Text className="mt-1" color="muted">
                Manage teams in your workspace.
              </Text>
            </Box>
            <Button color="primary" variant="outline">
              Create Team
            </Button>
          </Flex>
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
              <Flex align="center" gap={3}>
                <Flex className="-space-x-2">
                  <Avatar
                    className="h-6 w-6 ring-2 ring-white dark:ring-dark-100"
                    name="John Doe"
                  />
                  <Avatar
                    className="h-6 w-6 ring-2 ring-white dark:ring-dark-100"
                    name="Jane Smith"
                  />
                  <Avatar
                    className="h-6 w-6 ring-2 ring-white dark:ring-dark-100"
                    name="Bob Wilson"
                  />
                </Flex>
                <Button color="tertiary" variant="outline">
                  Manage
                </Button>
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
              <Flex align="center" gap={3}>
                <Flex className="-space-x-2">
                  <Avatar
                    className="h-6 w-6 ring-2 ring-white dark:ring-dark-100"
                    name="Alice Johnson"
                  />
                  <Avatar
                    className="h-6 w-6 ring-2 ring-white dark:ring-dark-100"
                    name="Charlie Brown"
                  />
                </Flex>
                <Button color="tertiary" variant="outline">
                  Manage
                </Button>
              </Flex>
            </Flex>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
