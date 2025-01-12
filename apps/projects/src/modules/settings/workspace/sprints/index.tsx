"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../components";

export const SprintsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Sprint Settings
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Configure default settings for sprints in your workspace."
          title="Sprint Configuration"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Sprint Duration</Text>
              <Text className="mt-1" color="muted">
                Set the default duration for new sprints.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Duration
              </Button>
            </Box>

            <Box>
              <Text className="font-medium">Sprint Capacity</Text>
              <Text className="mt-1" color="muted">
                Configure how story points and work items are tracked.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Capacity
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button color="primary" variant="outline">
              Create Template
            </Button>
          }
          description="Create and manage sprint templates for different types of work."
          title="Sprint Templates"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No sprint templates available.</Text>
              <Text className="mt-2" color="muted">
                Templates help you quickly set up sprints with predefined
                settings.
              </Text>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Configure sprint metrics and reporting settings."
          title="Sprint Analytics"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Velocity Tracking</Text>
              <Text className="mt-1" color="muted">
                Configure how team velocity is calculated and displayed.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Metrics
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
