"use client";

import { useId } from "react";
import { Box, Button, Flex, Switch, Text } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useDesktopPreferences } from "./use-desktop-preferences";
import { DesktopUpdateSettings } from "./desktop-update-settings";
import { DesktopCacheStatus } from "./desktop-cache-status";
import { DesktopOutboxStatus } from "./desktop-outbox-status";

export const DesktopSettings = ({
  section,
}: {
  section: "behavior" | "notifications";
}) => {
  const id = useId();
  const { supported, preferences, update } = useDesktopPreferences();
  if (!supported) return null;
  const error = update.error ?? preferences.error;
  const errorMessage = error instanceof Error ? error.message : error;
  const settings = preferences.data;
  const rows =
    section === "notifications"
      ? [
          {
            key: "notifications" as const,
            title: "Desktop notifications",
            description:
              "Show system notifications for new activity in this workspace while FortyOne is running.",
          },
        ]
      : [
          {
            key: "keepRunning" as const,
            title: "Keep running when the window closes",
            description:
              "Keep this workspace connected and receive notifications in the background. Quitting FortyOne stops notifications.",
          },
          {
            key: "autostart" as const,
            title: "Launch at login",
            description:
              "Open FortyOne automatically when you sign in to your computer.",
          },
        ];

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description={
          section === "notifications"
            ? "Manage notifications on this computer."
            : "Choose how FortyOne runs on this computer."
        }
        title={
          section === "notifications" ? "Desktop Notifications" : "Desktop App"
        }
      />
      <Flex className="p-6" direction="column" gap={6}>
        {!settings && preferences.isPending ? (
          <Text color="muted" role="status">
            Loading desktop preferences…
          </Text>
        ) : null}
        {settings
          ? rows.map(({ key, title, description }) => (
              <Flex align="center" gap={4} justify="between" key={key}>
                <Box>
                  <Text className="font-medium" id={`${id}-${key}-label`}>
                    {title}
                  </Text>
                  <Text color="muted" id={`${id}-${key}-description`}>
                    {description}
                  </Text>
                </Box>
                <Switch
                  aria-describedby={`${id}-${key}-description`}
                  aria-labelledby={`${id}-${key}-label`}
                  checked={settings[key]}
                  className="shrink-0"
                  disabled={update.isPending}
                  onCheckedChange={(checked) => {
                    update.mutate({ [key]: checked });
                  }}
                />
              </Flex>
            ))
          : null}
        {update.isPending ? (
          <Text color="muted" role="status">
            Saving desktop preference…
          </Text>
        ) : null}
        {error ? (
          <Box role="alert">
            <Text>
              {typeof errorMessage === "string"
                ? errorMessage
                : "Could not update desktop preferences. Try again."}
            </Text>
            {!settings ? (
              <Button
                className="mt-3"
                onClick={() => {
                  void preferences.refetch();
                }}
                variant="outline"
              >
                Retry
              </Button>
            ) : null}
          </Box>
        ) : null}
        {section === "behavior" ? (
          <>
            <DesktopCacheStatus />
            <DesktopOutboxStatus />
            <DesktopUpdateSettings />
          </>
        ) : null}
      </Flex>
    </Box>
  );
};
