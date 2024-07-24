"use client";
import { Button, Container, Dialog, Flex, Text, Tooltip } from "ui";
import { CopyIcon, DeleteIcon, LinkIcon } from "icons";
import { DetailedStory } from "../types";
import { useStore } from "@/hooks/store";
import { useCopyToClipboard } from "@/hooks";
import { toast } from "sonner";
import nProgress from "nprogress";
import { deleteStoryAction } from "../actions/delete-story";
import { useState } from "react";
import { restoreStoryAction } from "@/modules/story/actions/restore-story";

export const OptionsHeader = ({ story }: { story: DetailedStory }) => {
  const { id, sequenceId, deletedAt } = story;
  const [loading, setLoading] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const { teams } = useStore();
  const [_, copyText] = useCopyToClipboard();
  const { code } = teams.find((team) => team.id === story.teamId)!!;
  const isDeleted = !!deletedAt;

  const restoreStory = async () => {
    try {
      nProgress.start();
      const _ = await restoreStoryAction(id);
      toast.success("Success", {
        description: "Story restored successfully",
      });
    } catch (error) {
      toast.error("Error", {
        description: "Failed to restore story",
      });
    } finally {
      nProgress.done();
    }
  };

  const handleDelete = async () => {
    try {
      nProgress.start();
      setLoading(true);
      await deleteStoryAction(id);
      toast.success("Success", {
        description: "Story deleted successfully",
        cancel: {
          label: "Undo",
          onClick: restoreStory,
        },
      });
    } catch (e) {
      toast.error("Error", {
        description: "Failed to delete story",
      });
    } finally {
      nProgress.done();
      setLoading(false);
      setIsOpen(false);
    }
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
              leftIcon={<LinkIcon className="h-5 w-auto" strokeWidth={2.5} />}
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
          <Tooltip hidden={isDeleted} title="Delete Story">
            <Button
              color="danger"
              onClick={() => {
                setIsOpen(true);
              }}
              disabled={isDeleted}
              leftIcon={<DeleteIcon className="h-5 w-auto" />}
              variant="naked"
            >
              <span className="sr-only">Delete story</span>
            </Button>
          </Tooltip>
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
                leftIcon={<DeleteIcon className="h-5 w-auto" />}
                loading={loading}
                loadingText="Deleting..."
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
