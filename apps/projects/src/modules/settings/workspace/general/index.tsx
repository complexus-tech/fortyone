"use client";

import { Box, Flex, Text, Input } from "ui";
import { SectionHeader } from "../../components";

export const WorkspaceGeneralSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Workspace Settings
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Basic information about your workspace."
          title="General Information"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Input
                helpText="This is your workspace's visible name in Complexus"
                label="Workspace name"
                name="name"
                placeholder="Enter workspace name"
              />
            </Box>

            <Box>
              <Input
                helpText="This is your workspace's URL on Complexus"
                label="Workspace URL"
                name="url"
                placeholder="your-workspace"
                prefix="complexus.tech/"
              />
            </Box>

            <Box>
              <Input
                helpText="Brief description of your workspace"
                label="Description"
                name="description"
                placeholder="Enter workspace description"
              />
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
