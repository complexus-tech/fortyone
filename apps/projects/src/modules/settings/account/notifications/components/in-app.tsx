import React from "react";
import { Box, Flex } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useNotificationPreferences } from "@/modules/notifications/hooks/preferences";
import { Entry } from "./entry";

export const InAppNotifications = () => {
  const { data } = useNotificationPreferences();
  const preferences = data?.preferences;

  return (
    <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description="Choose what updates you want to receive via in-app notifications."
        title="In-App Notifications"
      />
      <Box className="p-6">
        <Flex direction="column" gap={6}>
          <Entry
            checked={preferences?.story_update.inApp}
            description="Get notified when a story you're involved with is updated"
            title="Story updates"
          />

          <Entry
            checked={preferences?.comment_reply.inApp}
            description="Get notified when someone comments on your stories"
            title="Comments"
          />

          <Entry
            checked={preferences?.mention.inApp}
            description="Get notified when someone mentions you in a comment or story"
            title="Mentions"
          />
        </Flex>
      </Box>
    </Box>
  );
};
