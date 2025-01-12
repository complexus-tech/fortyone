"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../../components";

export const IntegrationsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Integrations
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Connect with your favorite tools and services."
          title="Available Integrations"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Development Tools</Text>
              <Text className="mt-1" color="muted">
                Connect with your development workflow:
              </Text>
              <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
                <li>GitHub</li>
                <li>GitLab</li>
                <li>Bitbucket</li>
                <li>VS Code</li>
              </ul>
              <Button className="mt-4" color="tertiary" variant="outline">
                Browse Development Tools
              </Button>
            </Box>

            <Box>
              <Text className="font-medium">Productivity Tools</Text>
              <Text className="mt-1" color="muted">
                Enhance your workflow with these integrations:
              </Text>
              <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
                <li>Slack</li>
                <li>Microsoft Teams</li>
                <li>Google Workspace</li>
                <li>Zoom</li>
              </ul>
              <Button className="mt-4" color="tertiary" variant="outline">
                Browse Productivity Tools
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Manage your connected integrations and their permissions."
          title="Connected Services"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No services connected yet.</Text>
              <Button className="mt-4" color="primary">
                Connect a Service
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
