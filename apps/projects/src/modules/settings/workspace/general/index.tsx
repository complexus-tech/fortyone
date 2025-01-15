"use client";

import { Box, Text, Button, Avatar } from "ui";
import { SectionHeader } from "../../components";
import { WorkspaceForm } from "./components/form";

export const WorkspaceGeneralSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Workspace Settings
      </Text>

      <Box className="mb-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={<Avatar name="Complexus Tech" />}
          description="Basic information about your workspace."
          title="General Information"
        />
        <WorkspaceForm />
      </Box>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Permanently delete your workspace."
          title="Delete your workspace"
        />

        <Box className="p-6">
          <Text className="mb-4" color="muted">
            Once you delete your workspace, there is no going back. Please be
            certain. All data will be lost including all teams, stories, and
            more.
          </Text>
          <Button
            className="mt-4 bg-primary dark:bg-primary"
            disabled
            variant="naked"
          >
            Delete Workspace
          </Button>
        </Box>
      </Box>
    </Box>
  );
};
