import React from "react";
import { Box, Flex } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useNotificationPreferences } from "@/modules/notifications/hooks/preferences";
import { useUpdateNotificationPreferenceMutation } from "@/modules/notifications/hooks/update-preference-mutation";
import { Entry } from "./entry";

type NotificationType =
  | "story_update"
  | "objective_update"
  | "comment_reply"
  | "mention"
  | "key_result_update"
  | "story_comment";

export const InAppNotifications = () => {
  const { data } = useNotificationPreferences();
  const preferences = data?.preferences;
  const { mutate } = useUpdateNotificationPreferenceMutation();

  const handleTogglePreference = (type: NotificationType, checked: boolean) => {
    mutate({
      type,
      preferences: {
        inAppEnabled: checked,
      },
    });
  };

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
            onChange={(checked) => {
              handleTogglePreference("story_update", checked);
            }}
            title="Story updates"
          />

          <Entry
            checked={preferences?.comment_reply.inApp}
            description="Get notified when someone comments on your stories"
            onChange={(checked) => {
              handleTogglePreference("comment_reply", checked);
            }}
            title="Comments"
          />

          <Entry
            checked={preferences?.mention.inApp}
            description="Get notified when someone mentions you in a comment or story"
            onChange={(checked) => {
              handleTogglePreference("mention", checked);
            }}
            title="Mentions"
          />
        </Flex>
      </Box>
    </Box>
  );
};
