"use client";

import { Box, Flex, Text, Switch, Select } from "ui";
import { SunIcon, MoonIcon, SystemIcon } from "icons";
import { useTheme } from "next-themes";
import { SectionHeader } from "../../components";

export const PreferencesSettings = () => {
  const { theme, setTheme } = useTheme();

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Preferences
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Choose what updates you want to receive via email."
          title="Email Notifications"
        />
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
        <SectionHeader
          description="Customize the application theme."
          title="Appearance"
        />

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

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Configure how stories are automatically handled."
          title="Automation"
        />
        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Flex align="center" justify="between">
              <Box>
                <Text className="font-medium">Auto-assign to self</Text>
                <Text color="muted">
                  When creating new stories, always assign them to yourself by
                  default
                </Text>
              </Box>
              <Switch name="autoAssignSelf" />
            </Flex>

            <Flex align="center" justify="between">
              <Box>
                <Text className="font-medium">
                  On git branch copy, move story to started status
                </Text>
                <Text color="muted">
                  After copying the git branch name, story is moved to the
                  started workflow status
                </Text>
              </Box>
              <Switch name="autoBranchMoveStatus" />
            </Flex>

            <Flex align="center" justify="between">
              <Box>
                <Text className="font-medium">
                  On git branch copy, assign to yourself
                </Text>
                <Text color="muted">
                  After copying the git branch name, story is assigned to
                  yourself
                </Text>
              </Box>
              <Switch name="autoBranchAssign" />
            </Flex>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
