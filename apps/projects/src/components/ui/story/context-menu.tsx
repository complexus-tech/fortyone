"use client";
import type { ReactNode } from "react";
import { Fragment, useState } from "react";
import { Box, Button, ContextMenu, Dialog, Text } from "ui";
import {
  ArchiveIcon,
  DeleteIcon,
  DuplicateIcon,
  EditIcon,
  ExternalLinkIcon,
  ShareIcon,
  StoryIcon,
  UndoIcon,
} from "icons";
import { useRouter, usePathname } from "next/navigation";
import { toast } from "sonner";
import type { Story } from "@/modules/stories/types";
import {
  useCopyToClipboard,
  useTerminology,
  useUserRole,
  useWorkspacePath,
} from "@/hooks";
import { getStoryPath } from "@/shared/routing/story";
import { useBulkDeleteStoryMutation } from "@/modules/stories/hooks/delete-mutation";
import { useBulkArchiveStoryMutation } from "@/modules/stories/hooks/archive-mutation";
import { useBulkUnarchiveStoryMutation } from "@/modules/stories/hooks/unarchive-mutation";
import { useBulkRestoreStoryMutation } from "@/modules/stories/hooks/restore-mutation";
import { useDuplicateStoryMutation } from "@/modules/story/hooks/duplicate-mutation";
import type { DetailedStory } from "@/modules/story/types";
import { ContextMenuItem } from "./context-menu-item";
import { StoryPropertyContextMenu } from "./story-property-context-menu";

const StoryActionConfirmationDialog = ({
  confirmIcon,
  confirmLabel,
  description,
  onConfirm,
  onOpenChange,
  open,
  title,
}: {
  confirmIcon?: ReactNode;
  confirmLabel: string;
  description: ReactNode;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
  open: boolean;
  title: string;
}) => (
  <Dialog onOpenChange={onOpenChange} open={open}>
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title className="px-6 pt-0.5 text-lg">{title}</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body>
        <Text color="muted">{description}</Text>
      </Dialog.Body>
      <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
        <Button
          className="px-4"
          color="tertiary"
          onClick={() => {
            onOpenChange(false);
          }}
        >
          Cancel
        </Button>
        <Button className="px-4" leftIcon={confirmIcon} onClick={onConfirm}>
          {confirmLabel}
        </Button>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog>
);

export const StoryContextMenu = ({
  children,
  story,
}: {
  children: ReactNode;
  story: Story;
}) => {
  const router = useRouter();
  const pathname = usePathname();
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isArchiveDialogOpen, setIsArchiveDialogOpen] = useState(false);
  const [isUnarchiveDialogOpen, setIsUnarchiveDialogOpen] = useState(false);
  const [isRestoreDialogOpen, setIsRestoreDialogOpen] = useState(false);
  const [_, copy] = useCopyToClipboard();
  const { withWorkspace } = useWorkspacePath();
  const { mutate: deleteStory } = useBulkDeleteStoryMutation();
  const { mutate: archiveStory } = useBulkArchiveStoryMutation();
  const { mutate: unarchiveStory } = useBulkUnarchiveStoryMutation();
  const { mutate: restoreStory } = useBulkRestoreStoryMutation();
  const { mutate: duplicateStory } = useDuplicateStoryMutation();
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");

  const isOnDeletedPage = pathname.includes("/deleted");
  const isOnArchivePage = pathname.includes("/archived");
  const canUpdateProperties = userRole !== "guest" && !isOnDeletedPage;

  const storyUrl = withWorkspace(
    getStoryPath({
      id: story.id,
      sequenceId: story.sequenceId,
      teamCode: story.team?.code,
    }),
  );

  const handleArchive = () => {
    archiveStory([story.id]);
    setIsArchiveDialogOpen(false);
  };

  const handleUnarchive = () => {
    unarchiveStory([story.id]);
    setIsUnarchiveDialogOpen(false);
  };

  const handleRestore = () => {
    restoreStory([story.id]);
    setIsRestoreDialogOpen(false);
  };

  const contextMenu = [
    {
      name: "Main",
      options: [
        {
          label: userRole === "guest" ? "View" : "Edit",
          icon: userRole === "guest" ? <StoryIcon /> : <EditIcon />,
          onSelect: () => {
            router.push(storyUrl);
          },
        },
        {
          label: "Duplicate",
          icon: <DuplicateIcon />,
          disabled: userRole === "guest",
          onSelect: () => {
            duplicateStory({
              storyId: story.id,
              story: {
                title: story.title,
                teamId: story.teamId,
                objectiveId: story.objectiveId,
                keyResultId: story.keyResultId,
                sprintId: story.sprintId,
                statusId: story.statusId,
                assigneeId: story.assigneeId,
                priority: story.priority,
                startDate: story.startDate,
                endDate: story.endDate,
              } as DetailedStory,
            });
          },
        },
        {
          label: "Open in new tab",
          icon: <ExternalLinkIcon className="h-[1.15rem]" />,
          onSelect: () => {
            window.open(storyUrl, "_blank", "noopener,noreferrer");
          },
        },
        {
          label: "Share link",
          icon: <ShareIcon />,
          onSelect: async () => {
            await copy(window.location.origin + storyUrl);
            toast.success("Link copied to clipboard");
          },
        },
        // Archive option - show only on active stories
        ...(!isOnDeletedPage && !isOnArchivePage
          ? [
              {
                label: `Archive ${storyTerm}`,
                disabled: userRole === "guest",
                icon: <ArchiveIcon />,
                onSelect: () => {
                  setIsArchiveDialogOpen(true);
                },
              },
            ]
          : []),
        // Unarchive option - show only on archived stories
        ...(isOnArchivePage
          ? [
              {
                label: `Unarchive ${storyTerm}`,
                disabled: userRole === "guest",
                icon: <ArchiveIcon />,
                onSelect: () => {
                  setIsUnarchiveDialogOpen(true);
                },
              },
            ]
          : []),
        // Restore option - show only on deleted stories
        ...(isOnDeletedPage
          ? [
              {
                label: `Restore ${storyTerm}`,
                disabled: userRole === "guest",
                icon: <UndoIcon />,
                onSelect: () => {
                  setIsRestoreDialogOpen(true);
                },
              },
            ]
          : []),
      ],
    },
    {
      name: "Danger Zone",
      options: [
        {
          label:
            isOnDeletedPage || isOnArchivePage ? "Delete forever" : "Delete",
          icon: <DeleteIcon className="text-danger dark:text-danger" />,
          disabled: userRole === "guest",
          onSelect: () => {
            setIsDeleteOpen(true);
          },
        },
      ],
    },
  ];

  return (
    <>
      <ContextMenu>
        <ContextMenu.Trigger>
          <Box>{children}</Box>
        </ContextMenu.Trigger>
        <ContextMenu.Items className="w-56">
          <StoryPropertyContextMenu
            disabled={!canUpdateProperties}
            story={story}
          />
          {contextMenu.map(({ name, options }) => (
            <Fragment key={name}>
              <ContextMenu.Group key={name}>
                {options.map(({ label, icon, onSelect, disabled }) => (
                  <ContextMenuItem
                    disabled={disabled}
                    icon={icon}
                    key={label}
                    label={label}
                    labelColor={name === "Danger Zone" ? "danger" : undefined}
                    onSelect={onSelect}
                  />
                ))}
              </ContextMenu.Group>
              {name !== "Danger Zone" && (
                <ContextMenu.Separator className="my-2" />
              )}
            </Fragment>
          ))}
        </ContextMenu.Items>
      </ContextMenu>

      <StoryActionConfirmationDialog
        confirmLabel={
          isOnDeletedPage || isOnArchivePage ? "Delete forever" : "Delete"
        }
        description={
          isOnDeletedPage || isOnArchivePage
            ? `This is an irreversible action. The ${storyTerm} will be permanently deleted. You can't restore it.`
            : `This ${storyTerm} will be moved to the recycle bin and will be permanently deleted after 30 days. You can restore it at any time before that.`
        }
        onConfirm={() => {
          deleteStory({
            storyIds: [story.id],
            hardDelete: isOnDeletedPage || isOnArchivePage,
          });
          setIsDeleteOpen(false);
        }}
        onOpenChange={setIsDeleteOpen}
        open={isDeleteOpen}
        title={`Delete ${storyTerm}`}
      />
      <StoryActionConfirmationDialog
        confirmIcon={<ArchiveIcon className="text-current" />}
        confirmLabel="Archive"
        description={`This ${storyTerm} will be moved to the archive and can be unarchived later. It won't appear in your active ${storyTerm} lists.`}
        onConfirm={handleArchive}
        onOpenChange={setIsArchiveDialogOpen}
        open={isArchiveDialogOpen}
        title={`Archive ${storyTerm}`}
      />
      <StoryActionConfirmationDialog
        confirmIcon={<ArchiveIcon className="text-current" />}
        confirmLabel="Unarchive"
        description={`This ${storyTerm} will be restored to your active ${storyTerm} list and can be assigned to sprints and team members again.`}
        onConfirm={handleUnarchive}
        onOpenChange={setIsUnarchiveDialogOpen}
        open={isUnarchiveDialogOpen}
        title={`Unarchive ${storyTerm}`}
      />
      <StoryActionConfirmationDialog
        confirmIcon={<UndoIcon className="text-current" />}
        confirmLabel="Restore"
        description={`This ${storyTerm} will be restored to your active ${storyTerm} list and can be assigned to sprints and team members again.`}
        onConfirm={handleRestore}
        onOpenChange={setIsRestoreDialogOpen}
        open={isRestoreDialogOpen}
        title={`Restore ${storyTerm}`}
      />
    </>
  );
};
