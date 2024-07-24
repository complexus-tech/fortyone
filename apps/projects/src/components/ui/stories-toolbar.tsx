"use client";
import { Button, Flex, Text, Tooltip, Dialog } from "ui";
import { CloseIcon, DeleteIcon, ObjectiveIcon, SprintsIcon } from "icons";
import { useBoard } from "./board-context";
import { bulkDeleteAction } from "@/modules/stories/actions/bulk-delete-stories";
import nProgress from "nprogress";
import { useState } from "react";
import { toast } from "sonner";
import { bulkRestoreAction } from "@/modules/stories/actions/bulk-restore-stories";

export const StoriesToolbar = () => {
  const [loading, setLoading] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);
  const { selectedStories, setSelectedStories } = useBoard();

  const bulkRestore = async (stories: string[]) => {
    try {
      nProgress.start();
      setIsRestoring(true);
      const { storyIds } = await bulkRestoreAction(stories);
      toast.success("Success", {
        description: `${storyIds.length} stories restored`,
      });
    } catch (error) {
      console.log(error);
      toast.error("Error", {
        description: "Failed to restore stories",
      });
    } finally {
      setIsRestoring(false);
      nProgress.done();
    }
  };

  const handleBulkDelete = async () => {
    try {
      nProgress.start();
      setLoading(true);
      const { storyIds } = await bulkDeleteAction(selectedStories);
      toast.success("Success", {
        description: `${selectedStories.length} stories deleted`,
        cancel: {
          label: "Undo",
          onClick: () => {
            bulkRestore(storyIds);
          },
        },
      });
    } catch (error) {
      toast.error("Error", {
        description: "Failed to delete stories",
      });
    } finally {
      setLoading(false);
      nProgress.done();
      setSelectedStories([]);
      setIsOpen(false);
    }
  };
  return (
    <>
      <Flex
        align="center"
        className="fixed bottom-8 left-1/2 right-1/2 z-50 w-max -translate-x-1/2 rounded-[0.55rem] border-[0.5px] border-gray-100 bg-white/60 px-2.5 py-2 shadow-lg shadow-dark/10 backdrop-blur dark:border-dark-50 dark:bg-dark-300/70 dark:shadow-dark/20"
        gap={2}
      >
        <Text
          as="span"
          className="ml-2 mr-4 flex items-center gap-1.5 px-1 opacity-80"
        >
          <Tooltip title="Clear selection">
            <Button
              color="tertiary"
              leftIcon={
                <CloseIcon className="relative h-4 w-auto" strokeWidth={3} />
              }
              onClick={() => {
                setSelectedStories([]);
              }}
              size="xs"
              variant="outline"
            >
              <span className="sr-only">Clear</span>
            </Button>
          </Tooltip>
          {selectedStories.length} selected
        </Text>
        <Button
          color="tertiary"
          leftIcon={<SprintsIcon className="h-[1.15rem] w-auto" />}
          variant="outline"
        >
          Add to sprint
        </Button>
        <Button
          color="tertiary"
          leftIcon={<ObjectiveIcon className="h-[1.15rem] w-auto" />}
          variant="outline"
        >
          Add to objective
        </Button>
        <Button
          leftIcon={<DeleteIcon className="h-[1.15rem] w-auto" />}
          onClick={() => {
            setIsOpen(true);
          }}
        >
          Delete
        </Button>
      </Flex>

      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content>
          <Dialog.Header className="flex items-center justify-between px-6 pt-6">
            <Dialog.Title className="flex items-center gap-1 text-lg">
              Delete {selectedStories.length} stories?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              These stories will be moved to the recycle bin and will be
              permanently deleted after 30 days. You can restore them at any
              time before that.
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
                loading={loading || isRestoring}
                loadingText={isRestoring ? "Restoring..." : "Deleting..."}
                onClick={handleBulkDelete}
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
