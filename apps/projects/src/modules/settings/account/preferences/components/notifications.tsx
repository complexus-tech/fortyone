import { Box, Flex, Text, Switch } from "ui";
import { SectionHeader } from "@/modules/settings/components";

export const Notifications = () => {
  return (
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
  );
};
