"use client";

import { Box, Text, Flex } from "ui";
import { Suspense } from "react";
import { SectionHeader } from "@/modules/settings/components";
import { EmailNotifications } from "./components/email";

export const NotificationsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Notifications
      </Text>

      <Suspense
        fallback={
          <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
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

      {/* <Tabs defaultValue="inApp">
        <Tabs.List className="mx-0 mb-3 md:mx-0">
          <Tabs.Tab leftIcon={<NotificationsUnreadIcon />} value="inApp">
            In-App
          </Tabs.Tab>
          <Tabs.Tab leftIcon={<EmailIcon />} value="email">
            Email
          </Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="email">
          <Suspense
            fallback={
              <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
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
        </Tabs.Panel>
        <Tabs.Panel value="inApp">
          <Suspense
            fallback={
              <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
                <SectionHeader
                  description="Choose what updates you want to receive via in-app notifications."
                  title="In-App Notifications"
                />
                <Box className="p-6">
                  <Flex direction="column" gap={6}>
                    <Text color="muted">Loading...</Text>
                  </Flex>
                </Box>
              </Box>
            }
          >
            <InAppNotifications />
          </Suspense>
        </Tabs.Panel>
      </Tabs> */}
    </Box>
  );
};
