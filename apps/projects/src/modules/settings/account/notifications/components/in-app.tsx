import React from "react";
import { Box, Flex } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useNotificationPreferences } from "@/modules/notifications/hooks/preferences";
import { useUpdateNotificationPreferenceMutation } from "@/modules/notifications/hooks/update-preference-mutation";
import { type NotificationType } from "@/modules/notifications/types";
import { notificationConfigs } from "../constants/notification-configs";
import { Entry } from "./entry";

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
          {notificationConfigs.map((config) => (
            <Entry
              checked={preferences?.[config.type]?.inApp}
              description={config.description}
              key={config.type}
              onChange={(checked) => {
                handleTogglePreference(config.type, checked);
              }}
              title={config.title}
            />
          ))}
        </Flex>
      </Box>
    </Box>
  );
};
