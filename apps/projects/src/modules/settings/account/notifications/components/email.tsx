import React from "react";
import { Box, Flex } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useNotificationPreferences } from "@/modules/notifications/hooks/preferences";
import { Entry } from "./entry";

export const EmailNotifications = () => {
  const { data } = useNotificationPreferences();
  const preferences = data?.preferences;

  return (
    <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description="Choose what updates you want to receive via email."
        title="Email Notifications"
      />
      <Box className="p-6">
        <Flex direction="column" gap={6}>
          <Entry
            checked={preferences?.story_update.email}
            description="Get notified when a story you're involved with is updated"
            title="Story updates"
          />

          <Entry
            checked={preferences?.comment_reply.email}
            description="Get notified when someone comments on your stories"
            title="Comments"
          />

          <Entry
            checked={preferences?.mention.email}
            description="Get notified when someone mentions you in a comment or story"
            title="Mentions"
          />
        </Flex>
      </Box>
    </Box>
  );
};
