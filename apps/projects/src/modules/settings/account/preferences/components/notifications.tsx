import { Box, Flex, Text, Switch, Tabs } from "ui";
import { EmailIcon, NotificationsUnreadIcon } from "icons";
import { SectionHeader } from "@/modules/settings/components";
import { useNotificationPreferences } from "@/modules/notifications/hooks/preferences";

export const Notifications = () => {
  const { data } = useNotificationPreferences();
  const preferences = data?.preferences;
  return (
    <Tabs defaultValue="inApp">
      <Tabs.List className="mx-0 mb-3">
        <Tabs.Tab leftIcon={<NotificationsUnreadIcon />} value="inApp">
          In-App
        </Tabs.Tab>
        <Tabs.Tab leftIcon={<EmailIcon />} value="email">
          Email
        </Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="email">
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
                    Get notified when a story you&apos;re involved with is
                    updated
                  </Text>
                </Box>
                <Switch checked={preferences?.story_update.email} />
              </Flex>

              <Flex align="center" justify="between">
                <Box>
                  <Text className="font-medium">Comments</Text>
                  <Text color="muted">
                    Get notified when someone comments on your stories
                  </Text>
                </Box>
                <Switch checked={preferences?.comment_reply.email} />
              </Flex>
              <Flex align="center" justify="between">
                <Box>
                  <Text className="font-medium">Mentions</Text>
                  <Text color="muted">
                    Get notified when someone mentions you in a comment or story
                  </Text>
                </Box>
                <Switch checked={preferences?.mention.email} />
              </Flex>
            </Flex>
          </Box>
        </Box>
      </Tabs.Panel>
      <Tabs.Panel value="inApp" />
    </Tabs>
  );
};
