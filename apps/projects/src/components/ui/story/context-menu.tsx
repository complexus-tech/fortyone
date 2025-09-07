"use client";
import type { ReactNode } from "react";
import { Fragment, useState } from "react";
import { Box, Button, ContextMenu, Dialog, Text } from "ui";
import {
  ArchiveIcon,
  DeleteIcon,
  DuplicateIcon,
  EditIcon,
  NewTabIcon,
  ShareIcon,
  StoryIcon,
  UndoIcon,
} from "icons";
import { useRouter, usePathname } from "next/navigation";
import { toast } from "sonner";
import type { Story } from "@/modules/stories/types";
import { useCopyToClipboard, useUserRole } from "@/hooks";
import { slugify } from "@/utils";
import { useBulkDeleteStoryMutation } from "@/modules/stories/hooks/delete-mutation";
import { useBulkArchiveStoryMutation } from "@/modules/stories/hooks/archive-mutation";
import { useBulkUnarchiveStoryMutation } from "@/modules/stories/hooks/unarchive-mutation";
import { useBulkRestoreStoryMutation } from "@/modules/stories/hooks/restore-mutation";
import { useDuplicateStoryMutation } from "@/modules/story/hooks/duplicate-mutation";
import type { DetailedStory } from "@/modules/story/types";
import { ContextMenuItem } from "./context-menu-item";

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
  const { mutate: deleteStory } = useBulkDeleteStoryMutation();
  const { mutate: archiveStory } = useBulkArchiveStoryMutation();
  const { mutate: unarchiveStory } = useBulkUnarchiveStoryMutation();
  const { mutate: restoreStory } = useBulkRestoreStoryMutation();
  const { mutate: duplicateStory } = useDuplicateStoryMutation();
  const { userRole } = useUserRole();

  const isOnDeletedPage = pathname.includes("/deleted");
  const isOnArchivePage = pathname.includes("/archived");

  const storyUrl = `/story/${story.id}/${slugify(story.title)}`;

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
          icon: <NewTabIcon />,
          onSelect: () => {
            window.open(storyUrl, "_blank");
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
                label: "Archive story",
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
                label: "Unarchive story",
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
                label: "Restore story",
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
          {contextMenu.map(({ name, options }) => (
            <Fragment key={name}>
              <ContextMenu.Group key={name}>
                {options.map(({ label, icon, onSelect, disabled }) => (
                  <ContextMenuItem
                    disabled={disabled}
                    icon={icon}
                    key={label}
                    label={label}
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

      <Dialog onOpenChange={setIsDeleteOpen} open={isDeleteOpen}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Delete story
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text color="muted">
              {isOnDeletedPage || isOnArchivePage
                ? "This is an irreversible action. The story will be permanently deleted. You can't restore it."
                : "This story will be moved to the recycle bin and will be permanently deleted after 30 days. You can restore it at any time before that."}
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              className="px-4"
              color="tertiary"
              onClick={() => {
                setIsDeleteOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button
              className="px-4"
              onClick={() => {
                deleteStory({
                  storyIds: [story.id],
                  hardDelete: isOnDeletedPage || isOnArchivePage,
                });
                setIsDeleteOpen(false);
              }}
            >
              {isOnDeletedPage || isOnArchivePage ? "Delete forever" : "Delete"}
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>

      {/* Archive Dialog */}
      <Dialog onOpenChange={setIsArchiveDialogOpen} open={isArchiveDialogOpen}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Archive story
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text color="muted">
              This story will be moved to the archive and can be unarchived
              later. It won&apos;t appear in your active story lists.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              className="px-4"
              color="tertiary"
              onClick={() => {
                setIsArchiveDialogOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button
              className="px-4"
              leftIcon={<ArchiveIcon className="text-white dark:text-white" />}
              onClick={handleArchive}
            >
              Archive
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>

      {/* Unarchive Dialog */}
      <Dialog
        onOpenChange={setIsUnarchiveDialogOpen}
        open={isUnarchiveDialogOpen}
      >
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Unarchive story
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text color="muted">
              This story will be restored to your active story list and can be
              assigned to sprints and team members again.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              className="px-4"
              color="tertiary"
              onClick={() => {
                setIsUnarchiveDialogOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button
              className="px-4"
              leftIcon={<ArchiveIcon className="text-white dark:text-white" />}
              onClick={handleUnarchive}
            >
              Unarchive
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>

      {/* Restore Dialog */}
      <Dialog onOpenChange={setIsRestoreDialogOpen} open={isRestoreDialogOpen}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Restore story
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text color="muted">
              This story will be restored to your active story list and can be
              assigned to sprints and team members again.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              className="px-4"
              color="tertiary"
              onClick={() => {
                setIsRestoreDialogOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button
              className="px-4"
              leftIcon={<UndoIcon className="text-white dark:text-white" />}
              onClick={handleRestore}
            >
              Restore
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
