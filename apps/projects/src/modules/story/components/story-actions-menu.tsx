"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import {
  ArchiveIcon,
  ArrowRightIcon,
  BellIcon,
  CheckIcon,
  DeleteIcon,
  DuplicateIcon,
  ExternalLinkIcon,
  MoreHorizontalIcon,
  NotificationsOffIcon,
  ShareIcon,
  UndoIcon,
} from "icons";
import { Button, Dialog, Flex, Menu, Text } from "ui";
import { toast } from "sonner";
import type { StoryPriority } from "@/modules/stories/types";
import { STORY_PRIORITIES } from "@/components/ui/story/story-action-options";
import {
  useCopyToClipboard,
  useTerminology,
  useUserRole,
  useWorkspacePath,
} from "@/hooks";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useStatuses } from "@/lib/hooks/statuses";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { PriorityIcon, StoryStatusIcon } from "@/components/ui";
import { ObjectiveKeyResultSubMenu } from "@/components/ui/story/objective-key-result-menu";
import { useBulkArchiveStoryMutation } from "@/modules/stories/hooks/archive-mutation";
import { useBulkDeleteStoryMutation } from "@/modules/stories/hooks/delete-mutation";
import { useBulkRestoreStoryMutation } from "@/modules/stories/hooks/restore-mutation";
import { useBulkUnarchiveStoryMutation } from "@/modules/stories/hooks/unarchive-mutation";
import { getStoryPath } from "../utils/story-url";
import { useSetStoryWatchingMutation } from "../hooks/collaboration-mutations";
import { useDuplicateStoryMutation } from "../hooks/duplicate-mutation";
import { useStoryById } from "../hooks/story";
import { useUpdateStoryMutation } from "../hooks/update-mutation";

interface StoryStatusSubMenuProps {
  disabled: boolean;
  statusId: string;
  storyId: string;
  teamId: string;
}

const StoryStatusSubMenu = ({
  disabled,
  statusId,
  storyId,
  teamId,
}: StoryStatusSubMenuProps) => {
  const { data: statuses = [] } = useStatuses();
  const updateMutation = useUpdateStoryMutation();
  const teamStatuses = statuses.filter((status) => status.teamId === teamId);

  return (
    <Menu.SubMenu>
      <Menu.SubTrigger className="justify-between" disabled={disabled}>
        <Flex align="center" gap={2}>
          <StoryStatusIcon statusId={statusId} />
          Status
        </Flex>
        <ArrowRightIcon
          className="text-text-muted h-3.5 w-auto"
          strokeWidth={2.8}
        />
      </Menu.SubTrigger>
      <Menu.SubItems className="min-w-52">
        <Menu.Group>
          {teamStatuses.map((status) => (
            <Menu.Item
              className="justify-between"
              key={status.id}
              onSelect={() => {
                updateMutation.mutate({
                  payload: { statusId: status.id },
                  storyId,
                });
              }}
            >
              <Flex align="center" gap={2}>
                <StoryStatusIcon statusId={status.id} />
                <span className="max-w-44 truncate">{status.name}</span>
              </Flex>
              {status.id === statusId ? (
                <CheckIcon className="h-4 w-auto" />
              ) : null}
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.SubItems>
    </Menu.SubMenu>
  );
};

interface StoryPrioritySubMenuProps {
  disabled: boolean;
  priority: StoryPriority;
  storyId: string;
}

const StoryPrioritySubMenu = ({
  disabled,
  priority,
  storyId,
}: StoryPrioritySubMenuProps) => {
  const updateMutation = useUpdateStoryMutation();

  return (
    <Menu.SubMenu>
      <Menu.SubTrigger className="justify-between" disabled={disabled}>
        <Flex align="center" gap={2}>
          <PriorityIcon priority={priority} />
          Priority
        </Flex>
        <ArrowRightIcon
          className="text-text-muted h-3.5 w-auto"
          strokeWidth={2.8}
        />
      </Menu.SubTrigger>
      <Menu.SubItems className="min-w-48">
        <Menu.Group>
          {STORY_PRIORITIES.map((option) => (
            <Menu.Item
              className="justify-between"
              key={option}
              onSelect={() => {
                updateMutation.mutate({
                  payload: { priority: option },
                  storyId,
                });
              }}
            >
              <Flex align="center" gap={2}>
                <PriorityIcon priority={option} />
                {option}
              </Flex>
              {option === priority ? (
                <CheckIcon className="h-4 w-auto" />
              ) : null}
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.SubItems>
    </Menu.SubMenu>
  );
};

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
  isAdminOrOwner?: boolean;
  storyId: string;
}

export const StoryActionsMenu = ({
  align = "end",
  isAdminOrOwner,
  storyId,
}: StoryActionsMenuProps) => {
  const { data } = useStoryById(storyId);
  const [_, copyText] = useCopyToClipboard();
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();
  const { withWorkspace } = useWorkspacePath();
  const { isAdminOrOwner: derivedIsAdminOrOwner } = useIsAdminOrOwner(
    data?.reporterId,
  );
  const subscriptionMutation = useSetStoryWatchingMutation();
  const updateMutation = useUpdateStoryMutation();
  const archiveMutation = useBulkArchiveStoryMutation();
  const deleteMutation = useBulkDeleteStoryMutation();
  const duplicateMutation = useDuplicateStoryMutation();
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
  const canUpdateProperties = canEdit && !isDeleted;
  const storyUrl = withWorkspace(
    getStoryPath({
      id: data.id,
      sequenceId: data.sequenceId,
      teamCode: data.teamCode,
    }),
  );

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

  const handleDuplicate = () => {
    duplicateMutation.mutate({ story: data, storyId });
  };

  const handleShare = async () => {
    await copyText(`${window.location.origin}${storyUrl}`);
    toast.success("Link copied to clipboard");
  };

  return (
    <>
      <Menu>
        <Menu.Button>
          <Button
            aria-label="More story actions"
            asIcon
            color="tertiary"
            leftIcon={<MoreHorizontalIcon className="h-5 w-auto" />}
            variant="naked"
          >
            <span className="sr-only">More story actions</span>
          </Button>
        </Menu.Button>
        <Menu.Items align={align} className="min-w-52">
          <Menu.Group>
            <StoryStatusSubMenu
              disabled={!canUpdateProperties}
              statusId={data.statusId}
              storyId={storyId}
              teamId={data.teamId}
            />
            <StoryPrioritySubMenu
              disabled={!canUpdateProperties}
              priority={data.priority}
              storyId={storyId}
            />
            <ObjectiveKeyResultSubMenu
              disabled={!canUpdateProperties}
              keyResultId={data.keyResultId}
              objectiveId={data.objectiveId}
              onChange={(payload) => {
                updateMutation.mutate({ payload, storyId });
              }}
              teamId={data.teamId}
            />
            <Menu.Item
              disabled={!canEdit || duplicateMutation.isPending}
              onSelect={handleDuplicate}
            >
              <DuplicateIcon />
              Duplicate
            </Menu.Item>
            <Menu.Item
              onSelect={() => {
                window.open(storyUrl, "_blank", "noopener,noreferrer");
              }}
            >
              <ExternalLinkIcon className="h-[1.15rem]" />
              Open in new tab
            </Menu.Item>
            <Menu.Item onSelect={handleShare}>
              <ShareIcon />
              Copy link
            </Menu.Item>
          </Menu.Group>
          <Menu.Separator className="my-1.5" />
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
            {lifecycleMenuItem}
          </Menu.Group>
          <Menu.Separator className="my-1.5" />
          <Menu.Group>
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
