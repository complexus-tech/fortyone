"use client";

import { Box, Flex, Text, Switch } from "ui";

export const PreferencesSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Preferences
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 p-6 dark:border-dark-100">
          <Text as="h3" className="font-medium">
            Email Notifications
          </Text>
          <Text className="mt-1" color="muted">
            Choose what updates you want to receive via email.
          </Text>
        </Box>
        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Flex align="center" justify="between">
              <Box>
                <Text className="font-medium">Story updates</Text>
                <Text color="muted">
                  Get notified when a story you&apos;re involved with is updated
                </Text>
              </Box>
              <Switch defaultChecked />
            </Flex>

            <Flex align="center" justify="between">
              <Box>
                <Text className="font-medium">Comments</Text>
                <Text color="muted">
                  Get notified when someone comments on your stories
                </Text>
              </Box>
              <Switch defaultChecked />
            </Flex>
            <Flex align="center" justify="between">
              <Box>
                <Text className="font-medium">Mentions</Text>
                <Text color="muted">
                  Get notified when someone mentions you in a comment or story
                </Text>
              </Box>
              <Switch defaultChecked />
            </Flex>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
