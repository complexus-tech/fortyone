"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../components";

export const WebhooksSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Webhooks
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button color="primary" variant="outline">
              Add Webhook
            </Button>
          }
          description="Configure webhook endpoints to receive real-time updates."
          title="Webhook Endpoints"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No webhook endpoints configured.</Text>
              <Text className="mt-2" color="muted">
                Webhooks allow you to receive notifications when specific events
                occur in your workspace.
              </Text>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="View webhook delivery history and manage failed deliveries."
          title="Webhook History"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No webhook history available.</Text>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Configure webhook security settings and authentication."
          title="Security Settings"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Webhook Secret</Text>
              <Text className="mt-1" color="muted">
                Use webhook secrets to verify that requests came from your
                workspace.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Generate Secret
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
