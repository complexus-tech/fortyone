"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../../components";

export const AutomationsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Automations
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button color="primary" variant="outline">
              Create Rule
            </Button>
          }
          description="Configure automated actions based on triggers and conditions."
          title="Automation Rules"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No automation rules configured.</Text>
              <Text className="mt-2" color="muted">
                Automate repetitive tasks and workflows in your workspace.
              </Text>
            </Box>

            <Box>
              <Text className="font-medium">Available Triggers</Text>
              <Text className="mt-1" color="muted">
                Events that can start an automation:
              </Text>
              <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
                <li>Story created or updated</li>
                <li>Comment added</li>
                <li>Status changed</li>
                <li>Label added or removed</li>
                <li>Due date approaching</li>
              </ul>
            </Box>

            <Box>
              <Text className="font-medium">Available Actions</Text>
              <Text className="mt-1" color="muted">
                Actions that can be automated:
              </Text>
              <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
                <li>Assign story</li>
                <li>Update status</li>
                <li>Add label</li>
                <li>Send notification</li>
                <li>Create subtask</li>
              </ul>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="View automation execution history and manage failed runs."
          title="Automation History"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No automation history available.</Text>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
