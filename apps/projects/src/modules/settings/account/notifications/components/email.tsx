import React from "react";
import { Box, Flex } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useNotificationPreferences } from "@/modules/notifications/hooks/preferences";
import { useUpdateNotificationPreferenceMutation } from "@/modules/notifications/hooks/update-preference-mutation";
import { type NotificationType } from "@/modules/notifications/types";
import { useNotificationConfigs } from "../hooks/use-notification-configs";
import { Entry } from "./entry";

export const EmailNotifications = () => {
  const { data } = useNotificationPreferences();

  const { mutate } = useUpdateNotificationPreferenceMutation();
  const notificationConfigs = useNotificationConfigs();

  const handleTogglePreference = (type: NotificationType, checked: boolean) => {
    mutate({
      type,
      preferences: {
        emailEnabled: checked,
      },
    });
  };

  return (
    <Box className="rounded-2xl border border-border bg-surface">
      <SectionHeader
        description="Choose what updates you want to receive via email."
        title="Email Notifications"
      />
      <Box className="p-6">
        <Flex direction="column" gap={6}>
          {notificationConfigs.map((config) => (
            <Entry
              checked={data?.preferences[config.type].email}
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
