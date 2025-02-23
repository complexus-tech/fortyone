"use client";

import { Box, Text, Button } from "ui";
import { useState } from "react";
import { ConfirmDialog } from "@/components/ui";
import { SectionHeader } from "../../components";

export const DeleteAccountSettings = () => {
  const [isOpen, setIsOpen] = useState(false);

  const handleDelete = () => {
    setIsOpen(false);
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Delete Account
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Once you delete your account, there is no going back. Please be certain."
          title="Delete your account"
        />
        <Box className="p-6">
          <Text className="mb-4" color="muted">
            Permanently delete your profile, remove all personal information,
            and withdraw from all teams and workspaces.
          </Text>
          <Button
            onClick={() => {
              setIsOpen(true);
            }}
          >
            Delete Account
          </Button>
        </Box>
      </Box>

      <ConfirmDialog
        confirmText="Delete Account"
        description="This action cannot be undone. This will permanently delete your account and remove your data from our servers."
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={handleDelete}
        title="Delete my account"
      />
    </Box>
  );
};
