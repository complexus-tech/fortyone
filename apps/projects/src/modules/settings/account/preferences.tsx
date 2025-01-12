"use client";

import { Box, Flex, Text, Switch, Select } from "ui";
import { SunIcon, MoonIcon, SystemIcon } from "icons";
import { useTheme } from "next-themes";

export const PreferencesSettings = () => {
  const { theme, setTheme } = useTheme();

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Preferences
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
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

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
          <Text as="h3" className="font-medium">
            Appearance
          </Text>
          <Text className="mt-1" color="muted">
            Customize the application theme.
          </Text>
        </Box>

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Flex align="center" justify="between">
              <Box>
                <Text className="font-medium">Theme</Text>
                <Text color="muted">Select your preferred theme</Text>
              </Box>
              <Select
                defaultValue={theme}
                onValueChange={(value) => {
                  setTheme(value);
                }}
                value={theme}
              >
                <Select.Trigger className="w-max min-w-36">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Group>
                    <Select.Option value="light">
                      <Flex align="center" gap={2}>
                        <SunIcon className="h-4" />
                        Day Mode
                      </Flex>
                    </Select.Option>
                    <Select.Option value="dark">
                      <Flex align="center" gap={2}>
                        <MoonIcon className="h-4" />
                        Night Mode
                      </Flex>
                    </Select.Option>
                    <Select.Option value="system">
                      <Flex align="center" gap={2}>
                        <SystemIcon className="h-4" />
                        System preference
                      </Flex>
                    </Select.Option>
                  </Select.Group>
                </Select.Content>
              </Select>
            </Flex>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
