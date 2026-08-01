"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import {
  ArchiveIcon,
  BellIcon,
  DeleteIcon,
  MoreVerticalIcon,
  NotificationsOffIcon,
  UndoIcon,
} from "icons";
import { Button, Dialog, Flex, Menu, Text } from "ui";
import { useTerminology, useUserRole } from "@/hooks";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { useBulkArchiveStoryMutation } from "@/modules/stories/hooks/archive-mutation";
import { useBulkDeleteStoryMutation } from "@/modules/stories/hooks/delete-mutation";
import { useBulkRestoreStoryMutation } from "@/modules/stories/hooks/restore-mutation";
import { useBulkUnarchiveStoryMutation } from "@/modules/stories/hooks/unarchive-mutation";
import { useSetStoryWatchingMutation } from "../hooks/collaboration-mutations";
import { useStoryById } from "../hooks/story";

interface ConfirmationDialogProps {
  confirmIcon: ReactNode;
  confirmLabel: string;
  description: string;
  isPending?: boolean;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
  open: boolean;
  title: string;
}

const ConfirmationDialog = ({
  confirmIcon,
  confirmLabel,
  description,
  isPending = false,
  onConfirm,
  onOpenChange,
  open,
  title,
}: ConfirmationDialogProps) => (
  <Dialog onOpenChange={onOpenChange} open={open}>
    <Dialog.Content>
      <Dialog.Header className="px-6 pt-6">
        <Dialog.Title className="text-lg">{title}</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body className="pt-0">
        <Text color="muted">{description}</Text>
        <Flex align="center" className="mt-4" gap={2} justify="end">
          <Button
            color="tertiary"
            onClick={() => {
              onOpenChange(false);
            }}
          >
            Cancel
          </Button>
          <Button
            disabled={isPending}
            leftIcon={confirmIcon}
            loading={isPending}
            loadingText={`${confirmLabel}...`}
            onClick={onConfirm}
          >
            {confirmLabel}
          </Button>
        </Flex>
      </Dialog.Body>
    </Dialog.Content>
  </Dialog>
);

interface StoryActionsMenuProps {
  align?: "center" | "end" | "start";
  buttonClassName?: string;
  isAdminOrOwner?: boolean;
  storyId: string;
}

export const StoryActionsMenu = ({
  align = "end",
  buttonClassName,
  isAdminOrOwner,
  storyId,
}: StoryActionsMenuProps) => {
  const { data } = useStoryById(storyId);
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();
  const { isAdminOrOwner: derivedIsAdminOrOwner } = useIsAdminOrOwner(
    data?.reporterId,
  );
  const subscriptionMutation = useSetStoryWatchingMutation();
  const archiveMutation = useBulkArchiveStoryMutation();
  const deleteMutation = useBulkDeleteStoryMutation();
  const restoreMutation = useBulkRestoreStoryMutation();
  const unarchiveMutation = useBulkUnarchiveStoryMutation();
  const [isArchiveOpen, setIsArchiveOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isRestoreOpen, setIsRestoreOpen] = useState(false);
  const [isUnarchiveOpen, setIsUnarchiveOpen] = useState(false);
  const canDelete = isAdminOrOwner ?? derivedIsAdminOrOwner;

  useHotkeys(
    "backspace, delete",
    () => {
      if (canDelete && data && !data.deletedAt) {
        openDialogAfterMenuClose(setIsDeleteOpen);
      }
    },
    [canDelete, data],
  );

  if (!data) {
    return null;
  }

  const storyTerm = getTermDisplay("storyTerm");
  const storyTermCapitalized = getTermDisplay("storyTerm", {
    capitalize: true,
  });
  const isArchived = Boolean(data.archivedAt);
  const isDeleted = Boolean(data.deletedAt);
  const canEdit = userRole !== "guest";

  let lifecycleMenuItem: ReactNode;
  if (isDeleted) {
    lifecycleMenuItem = (
      <Menu.Item
        disabled={!canDelete || restoreMutation.isPending}
        onSelect={() => {
          openDialogAfterMenuClose(setIsRestoreOpen);
        }}
      >
        <UndoIcon />
        Restore {storyTerm}
      </Menu.Item>
    );
  } else if (isArchived) {
    lifecycleMenuItem = (
      <Menu.Item
        disabled={!canEdit || unarchiveMutation.isPending}
        onSelect={() => {
          openDialogAfterMenuClose(setIsUnarchiveOpen);
        }}
      >
        <ArchiveIcon />
        Unarchive {storyTerm}
      </Menu.Item>
    );
  } else {
    lifecycleMenuItem = (
      <Menu.Item
        disabled={!canEdit || archiveMutation.isPending}
        onSelect={() => {
          openDialogAfterMenuClose(setIsArchiveOpen);
        }}
      >
        <ArchiveIcon />
        Archive {storyTerm}
      </Menu.Item>
    );
  }

  const handleArchive = () => {
    archiveMutation.mutate([storyId]);
    setIsArchiveOpen(false);
  };

  const handleDelete = () => {
    deleteMutation.mutate({
      hardDelete: isDeleted,
      storyIds: [storyId],
    });
    setIsDeleteOpen(false);
  };

  const handleRestore = () => {
    restoreMutation.mutate([storyId]);
    setIsRestoreOpen(false);
  };

  const handleUnarchive = () => {
    unarchiveMutation.mutate([storyId]);
    setIsUnarchiveOpen(false);
  };

  return (
    <>
      <Menu>
        <Menu.Button>
          <Button
            aria-label="More story actions"
            asIcon
            className={buttonClassName}
            color="tertiary"
            leftIcon={<MoreVerticalIcon className="h-5 w-auto" />}
            variant="naked"
          >
            <span className="sr-only">More story actions</span>
          </Button>
        </Menu.Button>
        <Menu.Items align={align} className="min-w-52">
          <Menu.Group>
            <Menu.Item
              disabled={isDeleted || subscriptionMutation.isPending}
              onSelect={() => {
                subscriptionMutation.mutate({
                  storyId,
                  watching: !data.isWatching,
                });
              }}
            >
              {data.isWatching ? <NotificationsOffIcon /> : <BellIcon />}
              {data.isWatching ? "Unsubscribe" : "Subscribe"}
            </Menu.Item>
          </Menu.Group>
          <Menu.Separator className="my-1.5" />
          <Menu.Group>
            {lifecycleMenuItem}
            <Menu.Item
              className="text-danger"
              disabled={!canDelete || deleteMutation.isPending}
              onSelect={() => {
                openDialogAfterMenuClose(setIsDeleteOpen);
              }}
            >
              <DeleteIcon className="text-danger" />
              {isDeleted ? "Delete forever" : `Delete ${storyTerm}`}
            </Menu.Item>
          </Menu.Group>
        </Menu.Items>
      </Menu>

      <ConfirmationDialog
        confirmIcon={<ArchiveIcon className="text-white dark:text-gray-200" />}
        confirmLabel="Archive"
        description={`This ${storyTerm} will be moved to the archive and can be unarchived later.`}
        isPending={archiveMutation.isPending}
        onConfirm={handleArchive}
        onOpenChange={setIsArchiveOpen}
        open={isArchiveOpen}
        title={`Archive this ${storyTerm}?`}
      />
      <ConfirmationDialog
        confirmIcon={<ArchiveIcon className="text-white dark:text-gray-200" />}
        confirmLabel="Unarchive"
        description={`This ${storyTerm} will be restored to the active ${storyTerm} list.`}
        isPending={unarchiveMutation.isPending}
        onConfirm={handleUnarchive}
        onOpenChange={setIsUnarchiveOpen}
        open={isUnarchiveOpen}
        title={`Unarchive this ${storyTerm}?`}
      />
      <ConfirmationDialog
        confirmIcon={<UndoIcon className="text-white dark:text-gray-200" />}
        confirmLabel="Restore"
        description={`This ${storyTerm} will be restored from the recycle bin.`}
        isPending={restoreMutation.isPending}
        onConfirm={handleRestore}
        onOpenChange={setIsRestoreOpen}
        open={isRestoreOpen}
        title={`Restore this ${storyTerm}?`}
      />
      <ConfirmationDialog
        confirmIcon={<DeleteIcon className="text-white dark:text-gray-200" />}
        confirmLabel={isDeleted ? "Delete forever" : "Delete"}
        description={
          isDeleted
            ? `This ${storyTerm} will be permanently deleted. You can't restore it.`
            : `This ${storyTerm} will be moved to the recycle bin and permanently deleted after 30 days.`
        }
        isPending={deleteMutation.isPending}
        onConfirm={handleDelete}
        onOpenChange={setIsDeleteOpen}
        open={isDeleteOpen}
        title={`Delete this ${storyTermCapitalized}?`}
      />
    </>
  );
};
