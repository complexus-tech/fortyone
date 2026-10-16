"use client";

import type { ReactNode } from "react";
import { Suspense } from "react";
import { EmailIcon, NotificationsUnreadIcon } from "icons";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { Box, Flex, Tabs, Text } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { EmailNotifications } from "./components/email";

const NOTIFICATION_TABS = ["email", "desktop"] as const;

const EmailNotificationsPanel = () => (
  <Suspense
    fallback={
      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          description="Choose what updates you want to receive via email."
          title="Email Notifications"
        />
        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Text color="muted">Loading...</Text>
          </Flex>
        </Box>
      </Box>
    }
  >
    <EmailNotifications />
  </Suspense>
);

const NotificationTabs = ({
  desktopNotifications,
}: {
  desktopNotifications: ReactNode;
}) => {
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(NOTIFICATION_TABS)
      .withDefault("email")
      .withOptions({ history: "push" }),
  );

  return (
    <Tabs
      onValueChange={(value) => {
        if (value === "email" || value === "desktop") void setTab(value);
      }}
      value={tab}
    >
      <Box className="overflow-x-auto">
        <Tabs.List
          aria-label="Notification settings"
          className="mx-0 flex-nowrap md:mx-0"
        >
          <Tabs.Tab
            leftIcon={<EmailIcon className="h-[1.1rem]" />}
            value="email"
          >
            Email
          </Tabs.Tab>
          <Tabs.Tab
            leftIcon={<NotificationsUnreadIcon className="h-[1.1rem]" />}
            value="desktop"
          >
            Desktop
          </Tabs.Tab>
        </Tabs.List>
      </Box>
      <Box className="mt-5">
        <Tabs.Panel value="email">
          <EmailNotificationsPanel />
        </Tabs.Panel>
        <Tabs.Panel value="desktop">{desktopNotifications}</Tabs.Panel>
      </Box>
    </Tabs>
  );
};

export const NotificationsSettings = ({
  desktopNotifications,
}: {
  desktopNotifications?: ReactNode;
}) => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Notifications
      </Text>

      {desktopNotifications ? (
        <NotificationTabs desktopNotifications={desktopNotifications} />
      ) : (
        <EmailNotificationsPanel />
      )}
    </Box>
  );
};
