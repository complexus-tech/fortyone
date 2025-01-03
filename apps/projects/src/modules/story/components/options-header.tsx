"use client";
import { Button, Container, Dialog, Flex, Text, Tooltip } from "ui";
import { CopyIcon, DeleteIcon, LinkIcon, UndoIcon } from "icons";
import { useCopyToClipboard } from "@/hooks";
import { toast } from "sonner";
import { useState } from "react";
import { useParams } from "next/navigation";
import { useStoryById } from "@/modules/story/hooks/story";
import { useDeleteStoryMutation } from "../hooks/delete-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useRestoreStoryMutation } from "@/modules/story/hooks/restore-mutation";

export const OptionsHeader = () => {
  const params = useParams<{ storyId: string }>();
  const { data } = useStoryById(params.storyId);
  const { id, teamId, sequenceId, deletedAt } = data!;
  const [isOpen, setIsOpen] = useState(false);
  const { data: teams = [] } = useTeams();
  const [_, copyText] = useCopyToClipboard();
  const { code } = teams.find((team) => team.id === teamId)!!;
  const isDeleted = !!deletedAt;
  const { mutateAsync: deleteAsync } = useDeleteStoryMutation();
  const { mutateAsync } = useRestoreStoryMutation();

  const handleDelete = async () => {
    try {
      await deleteAsync(id);
    } finally {
      setIsOpen(false);
    }
  };

  const restoreStory = async () => {
    mutateAsync(id);
  };
  return (
    <>
      <Container className="flex h-16 w-full items-center justify-between px-7">
        <Text color="muted" transform="uppercase" fontWeight="semibold">
          {code}-{sequenceId}
        </Text>
        <Flex gap={2}>
          <Tooltip title="Copy Story Link">
            <Button
              color="tertiary"
              suppressHydrationWarning
              leftIcon={
                <LinkIcon className="h-5 w-auto -rotate-45" strokeWidth={2.5} />
              }
              variant="naked"
              onClick={async () => {
                await copyText(window.location.href);
                toast.info("Success", {
                  description: "Link copied to clipboard",
                });
              }}
            >
              <span className="sr-only">Copy story link</span>
            </Button>
          </Tooltip>
          <Tooltip title="Copy Story ID">
            <Button
              color="tertiary"
              leftIcon={<CopyIcon className="h-5 w-auto" />}
              variant="naked"
              onClick={async () => {
                await copyText(id);
                toast.info("Success", {
                  description: "Story ID copied to clipboard",
                });
              }}
            >
              <span className="sr-only">Copy story id</span>
            </Button>
          </Tooltip>
          {isDeleted ? (
            <Tooltip title="Restore Story">
              <Button
                color="tertiary"
                onClick={restoreStory}
                leftIcon={<UndoIcon />}
                variant="naked"
              >
                <span className="sr-only">Restore story</span>
              </Button>
            </Tooltip>
          ) : (
            <Tooltip title="Delete Story">
              <Button
                color="tertiary"
                onClick={() => {
                  setIsOpen(true);
                }}
                leftIcon={<DeleteIcon />}
                variant="naked"
              >
                <span className="sr-only">Delete story</span>
              </Button>
            </Tooltip>
          )}
        </Flex>
      </Container>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content>
          <Dialog.Header className="flex items-center justify-between px-6 pt-6">
            <Dialog.Title className="flex items-center gap-1 text-lg">
              Are you sure you want to delete this story?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              This story will be moved to the recycle bin and will be
              permanently deleted after 30 days. You can restore it at any time
              before that.
            </Text>
            <Flex align="center" gap={2} justify="end" className="mt-4">
              <Button
                color="tertiary"
                onClick={() => {
                  setIsOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                leftIcon={
                  <DeleteIcon className="text-white dark:text-gray-200" />
                }
                onClick={handleDelete}
              >
                Delete
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
