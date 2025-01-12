"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../components";

export const TemplatesSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Templates
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button color="primary" variant="outline">
              Create Template
            </Button>
          }
          description="Create and manage templates for different types of stories."
          title="Story Templates"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No story templates configured.</Text>
              <Text className="mt-2" color="muted">
                Templates help you create consistent stories with predefined
                fields and content.
              </Text>
            </Box>

            <Box>
              <Text className="font-medium">Template Components</Text>
              <Text className="mt-1" color="muted">
                Available components for templates:
              </Text>
              <ul className="mt-2 list-disc pl-5 text-dark/80 dark:text-[#949496]">
                <li>Title patterns</li>
                <li>Description templates</li>
                <li>Default fields</li>
                <li>Checklists</li>
                <li>Attachments</li>
              </ul>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Create templates for common workflows and processes."
          title="Workflow Templates"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Development Workflow</Text>
              <Text className="mt-1" color="muted">
                Standard workflow for software development tasks.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Template
              </Button>
            </Box>

            <Box>
              <Text className="font-medium">Bug Report Workflow</Text>
              <Text className="mt-1" color="muted">
                Workflow template for bug tracking and resolution.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Template
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Configure default settings for new templates."
          title="Template Settings"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Default Permissions</Text>
              <Text className="mt-1" color="muted">
                Configure who can create and edit templates.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Permissions
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
