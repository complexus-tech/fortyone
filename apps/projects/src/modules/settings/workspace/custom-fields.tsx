"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../components";

export const CustomFieldsSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Custom Fields
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button color="primary" variant="outline">
              Add Field
            </Button>
          }
          description="Create and manage custom fields for stories and tasks."
          title="Field Management"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No custom fields configured.</Text>
              <Text className="mt-2" color="muted">
                Add custom fields to track additional information in your
                stories.
              </Text>
            </Box>

            <Box>
              <Text className="font-medium">Available Field Types</Text>
              <Text className="mt-1" color="muted">
                Choose from these field types:
              </Text>
              <ul className="mt-2 list-disc pl-5 text-dark/80 dark:text-[#949496]">
                <li>Text (single line)</li>
                <li>Text area (multi-line)</li>
                <li>Number</li>
                <li>Date</li>
                <li>Single select</li>
                <li>Multi select</li>
                <li>User</li>
                <li>URL</li>
              </ul>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Configure default values and validation rules for custom fields."
          title="Field Configuration"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text className="font-medium">Default Values</Text>
              <Text className="mt-1" color="muted">
                Set default values for custom fields when creating new stories.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Defaults
              </Button>
            </Box>

            <Box>
              <Text className="font-medium">Required Fields</Text>
              <Text className="mt-1" color="muted">
                Specify which custom fields are required when creating or
                updating stories.
              </Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Configure Requirements
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
