"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../../components";

export const ImportsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Imports & Sync
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button color="primary" variant="outline">
              Start Import
            </Button>
          }
          description="Import data from other tools and services."
          title="Import Data"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Supported Import Sources</Text>
              <Text className="mt-1" color="muted">
                Import your data from these platforms:
              </Text>
              <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
                <li>Jira</li>
                <li>Trello</li>
                <li>Asana</li>
                <li>GitHub Issues</li>
                <li>CSV Files</li>
              </ul>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Configure automatic synchronization with external services."
          title="Sync Settings"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No sync connections configured.</Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Sync
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="View your past imports and their status."
          title="Import History"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No import history available.</Text>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
