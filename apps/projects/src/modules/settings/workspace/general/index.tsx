"use client";

import { Box, Text, Button } from "ui";
import { useState } from "react";
import { toast } from "sonner";
import { ConfirmDialog } from "@/components/ui";
import { logOut } from "@/components/shared/sidebar/actions";
import { clearAllStorage } from "@/components/shared/sidebar/utils";
import { useDeleteWorkspaceMutation } from "@/lib/hooks/delete-workspace-mutation";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useAnalytics } from "@/hooks";
import { SectionHeader } from "../../components";
import { WorkspaceForm } from "./components/form";
import { WorkspaceFeatures } from "./components/features";
import { Logo } from "./components/logo";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const WorkspaceGeneralSettings = () => {
  const [isOpen, setIsOpen] = useState(false);
  const { workspace } = useCurrentWorkspace();
  const { analytics } = useAnalytics();
  const { mutate: deleteWorkspace, isPending } = useDeleteWorkspaceMutation(
    workspace?.slug || "",
  );

  const handleDeleteWorkspace = async () => {
    try {
      await logOut();
      analytics.logout(true);
    } finally {
      clearAllStorage();
      window.location.href = `https://www.${domain}?signedOut=true`;
    }
  };

  const handleConfirmDelete = () => {
    deleteWorkspace(undefined, {
      onSuccess: () => {
        toast.success("Workspace deleted", {
          description: "Your workspace has been scheduled for deletion",
        });
        handleDeleteWorkspace();
      },
      onError: (error) => {
        toast.error("Failed to delete workspace", {
          description: error.message || "Please try again",
        });
      },
    });
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Workspace Settings
      </Text>
      <Box className="mb-6 rounded-2xl border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={<Logo />}
          description="Basic information about your workspace."
          title="General Information"
        />
        <WorkspaceForm />
      </Box>

      <Box className="mb-6">
        <WorkspaceFeatures />
      </Box>

      <Box className="rounded-2xl border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button
              className="mt-4 shrink-0"
              color="danger"
              disabled={isPending}
              onClick={() => {
                setIsOpen(true);
              }}
              variant="naked"
            >
              Delete <span className="hidden md:inline">Workspace</span>
            </Button>
          }
          description="Permanently delete your workspace."
          title="Danger Zone"
        />

        <Box className="p-6">
          <Text color="muted">
            Once you delete your workspace, there is no going back. Please be
            certain. All data will be lost including all teams, stories, and
            more.
          </Text>
        </Box>
      </Box>

      <ConfirmDialog
        confirmPhrase="delete workspace"
        confirmText="I understand"
        description="Once you delete your workspace, there is no going back. Your workspace will be scheduled for permanent deletion in 48 hours. All data will be lost including all teams, work items, and more."
        isLoading={isPending}
        isOpen={isOpen}
        loadingText="Deleting workspace..."
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={handleConfirmDelete}
        title="Delete Workspace"
      />
    </Box>
  );
};
