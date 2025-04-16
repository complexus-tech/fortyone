"use client";
import { Button, Flex, Text, Tooltip, Dialog, Command, Divider, Box } from "ui";
import { CloseIcon, DeleteIcon, SprintsIcon } from "icons";
import { useState } from "react";
import { useBulkDeleteStoryMutation } from "@/modules/stories/hooks/delete-mutation";
import { useSprints } from "@/modules/sprints/hooks/sprints";
import { useBoard } from "./board-context";

export const StoriesToolbar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const { selectedStories, setSelectedStories } = useBoard();

  const { data: sprints = [] } = useSprints();

  const { mutate: bulkDeleteMutate, isPending } = useBulkDeleteStoryMutation();

  const handleBulkDelete = () => {
    bulkDeleteMutate(selectedStories);
    setSelectedStories([]);
    setIsOpen(false);
  };

  return (
    <>
      <Flex
        align="center"
        className="fixed bottom-8 left-1/2 right-1/2 z-50 w-max -translate-x-1/2 rounded-xl border border-gray-100 bg-white/60 px-2.5 py-2 shadow-lg shadow-dark/10 backdrop-blur dark:border-dark-100 dark:bg-dark-300/70 dark:shadow-dark/20"
        gap={2}
      >
        <Text
          as="span"
          className="mr-4 flex items-center gap-1.5 px-1 opacity-80"
        >
          <Tooltip title="Clear selection">
            <Button
              color="tertiary"
              leftIcon={<CloseIcon className="h-4" strokeWidth={3} />}
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
        {/* <Button
          color="tertiary"
          disabled
          leftIcon={<SprintsIcon className="h-[1.15rem]" />}
          onClick={() => {
            setIsSprintsOpen(true);
          }}
          title="This is not available yet"
          variant="outline"
        >
          Add to sprint
        </Button>
        <Button
          color="tertiary"
          disabled
          leftIcon={<ObjectiveIcon className="h-[1.15rem]" />}
          onClick={() => {
            setIsObjectivesOpen(true);
          }}
          title="This is not available yet"
          variant="outline"
        >
          Add to objective
        </Button> */}
        <Button
          leftIcon={
            <DeleteIcon className="h-[1.15rem] text-white dark:text-gray-200" />
          }
          onClick={() => {
            setIsOpen(true);
          }}
        >
          Delete
        </Button>
      </Flex>

      {/* Delete stories dialog */}
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
            <Flex align="center" className="mt-4" gap={2} justify="end">
              <Button
                color="tertiary"
                disabled={isPending}
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
                loading={isPending}
                loadingText="Deleting stories..."
                onClick={handleBulkDelete}
              >
                Delete
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>

      {/* Add to sprint dialog */}
      <Dialog onOpenChange={setIsSprintsOpen} open={isSprintsOpen}>
        <Dialog.Content hideClose={false}>
          <Dialog.Header className="sr-only">
            <Dialog.Title>
              Add {selectedStories.length} stories to sprint
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="px-4 pt-4">
            <Command>
              <Command.Input
                autoFocus
                className="text-lg"
                placeholder="Choose sprint..."
              />
              <Divider className="mb-2 mt-3" />
              <Command.Empty className="py-2">
                <Text color="muted">No sprint found.</Text>
              </Command.Empty>
              <Command.Group className="px-0">
                {sprints.map(({ id, name }, idx) => (
                  <Command.Item
                    className="justify-between py-2.5"
                    key={id}
                    onSelect={() => {
                      setIsSprintsOpen(false);
                    }}
                    value={name}
                  >
                    <Box className="grid grid-cols-[24px_auto] items-center">
                      <SprintsIcon />
                      <Text>{name}</Text>
                    </Box>
                    <Flex align="center" gap={2}>
                      <Text color="muted">{idx}</Text>
                    </Flex>
                  </Command.Item>
                ))}
              </Command.Group>
            </Command>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>

      {/* Add to objective dialog */}
      <Dialog onOpenChange={setIsObjectivesOpen} open={isObjectivesOpen}>
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
            <Flex align="center" className="mt-4" gap={2} justify="end">
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
