"use client";

import { Box, Button, Text } from "ui";
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
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Profile Settings
      </Text>
      <Form />
      <Box className="mt-6 rounded-2xl border border-border bg-surface">
        <SectionHeader
          description="Leave your current workspace. This action cannot be undone."
          title="Leave Workspace"
        />
        <Box className="p-6">
          <Text className="mb-4" color="muted">
            You will lose access to all workspace data. Your account will remain
            active, and you can be invited back by workspace admins at any time.
          </Text>
          <Button
            color="danger"
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
        confirmPhrase="leave workspace"
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
