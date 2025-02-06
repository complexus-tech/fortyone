"use client";

import { Box, Button, Divider, Text } from "ui";
import { useState } from "react";
import { ConfirmDialog } from "@/components/ui";
import { SectionHeader } from "../../components";
import { Form } from "./components/form";

export const ProfileSettings = () => {
  const [isOpen, setIsOpen] = useState(false);

  const handleLeaveWorkspace = () => {
    setIsOpen(true);
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Profile Settings
      </Text>
      <Form />
      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Leave your current workspace. This action cannot be undone."
          title="Leave Workspace"
        />
        <Box className="p-6">
          <Box>
            <Text color="muted">When you leave the workspace:</Text>
            <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
              <li>You will lose access to all workspace data</li>
              <li>Your account will remain active</li>
              <li>You can be invited back by workspace admins</li>
            </ul>
          </Box>
          <Divider className="my-4" />
          <Button
            className="bg-primary dark:bg-primary"
            onClick={() => {
              setIsOpen(true);
            }}
            variant="naked"
          >
            Leave Workspace
          </Button>
        </Box>
      </Box>

      <ConfirmDialog
        confirmText="I understand"
        description="You will lose access to all workspace data. Your account will remain active."
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={handleLeaveWorkspace}
        title="Leave Workspace"
      />
    </Box>
  );
};
