"use client";

import { Box, Text, Button } from "ui";
import { useState } from "react";
import { ConfirmDialog } from "@/components/ui";
import { logOut } from "@/components/shared/sidebar/actions";
import { clearAllStorage } from "@/components/shared/sidebar/utils";
import { useDeleteWorkspaceMutation } from "@/lib/hooks/delete-workspace-mutation";
import { useRestoreWorkspaceMutation } from "@/lib/hooks/restore-workspace-mutation";
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
  const { mutate: deleteWorkspace, isPending: isDeleting } =
    useDeleteWorkspaceMutation();
  const { mutate: restoreWorkspace, isPending: isRestoring } =
    useRestoreWorkspaceMutation();

  const isWorkspaceDeleted = Boolean(workspace?.deletedAt);
  const isPending = isDeleting || isRestoring;

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
        handleDeleteWorkspace();
      },
    });
  };

  const handleRestore = () => {
    restoreWorkspace();
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Workspace Settings
      </Text>
      <Box className="mb-6 rounded-2xl border border-border bg-surface">
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

      <Box className="rounded-2xl border border-border bg-surface">
        <SectionHeader
          action={
            <Button
              className="mt-4 shrink-0"
              color="danger"
              loading={isPending}
              loadingText="Please wait..."
              onClick={() => {
                if (isWorkspaceDeleted) {
                  handleRestore();
                } else {
                  setIsOpen(true);
                }
              }}
              variant="naked"
            >
              {isWorkspaceDeleted ? "Restore" : "Delete"}{" "}
              <span className="hidden md:inline">Workspace</span>
            </Button>
          }
          description={
            isWorkspaceDeleted
              ? "Restore your deleted workspace."
              : "Permanently delete your workspace."
          }
          title={isWorkspaceDeleted ? "Restore Workspace" : "Danger Zone"}
        />

        <Box className="p-6">
          <Text color="muted">
            {isWorkspaceDeleted ? (
              <>
                Your workspace was deleted and will be permanently removed after
                the 48-hour grace period. You can restore it now to regain
                access to all your data.
              </>
            ) : (
              <>
                Once you delete your workspace, there is no going back. Please
                be certain. All data will be lost including all teams, work
                items, and more.
              </>
            )}
          </Text>
        </Box>
      </Box>
      <ConfirmDialog
        confirmPhrase="delete workspace"
        confirmText="I understand"
        description="Once you delete your workspace, there is no going back. Your workspace will be scheduled for permanent deletion in 48 hours. All data will be lost including all teams, work items, and more."
        isLoading={isDeleting}
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
