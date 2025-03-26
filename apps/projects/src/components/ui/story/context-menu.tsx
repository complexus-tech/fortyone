"use client";
import type { ReactNode } from "react";
import { Fragment, useState } from "react";
import { Box, Button, ContextMenu, Dialog, Text } from "ui";
import {
  DeleteIcon,
  DuplicateIcon,
  EditIcon,
  NewTabIcon,
  ShareIcon,
} from "icons";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import type { Story } from "@/modules/stories/types";
import { useCopyToClipboard } from "@/hooks";
import { slugify } from "@/utils";
import { useBulkDeleteStoryMutation } from "@/modules/stories/hooks/delete-mutation";
import { useDuplicateStoryMutation } from "@/modules/story/hooks/duplicate-mutation";
import { ContextMenuItem } from "./context-menu-item";

export const StoryContextMenu = ({
  children,
  story,
}: {
  children: ReactNode;
  story: Story;
}) => {
  const router = useRouter();
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [_, copy] = useCopyToClipboard();
  const { mutate: deleteStory } = useBulkDeleteStoryMutation();
  const { mutate: duplicateStory } = useDuplicateStoryMutation();

  const storyUrl = `/story/${story.id}/${slugify(story.title)}`;

  const contextMenu = [
    {
      name: "Main",
      options: [
        {
          label: "Edit",
          icon: <EditIcon />,
          onSelect: () => {
            router.push(storyUrl);
          },
        },
        {
          label: "Duplicate",
          icon: <DuplicateIcon />,
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
              },
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
      ],
    },
    {
      name: "Danger Zone",
      options: [
        {
          label: "Delete",
          icon: <DeleteIcon className="text-danger dark:text-danger" />,
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
                {options.map(({ label, icon, onSelect }) => (
                  <ContextMenuItem
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
              This story will be moved to the recycle bin and will be
              permanently deleted after 30 days. You can restore it at any time
              before that.
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
                deleteStory([story.id]);
                setIsDeleteOpen(false);
              }}
            >
              Delete
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
