"use client";
import { Button, Flex, Text, Tooltip, Dialog, Command, Divider, Box } from "ui";
import { CloseIcon, DeleteIcon, ObjectiveIcon, SprintsIcon } from "icons";
import { useBoard } from "./board-context";
import { useState } from "react";
import { useBulkDeleteStoryMutation } from "@/modules/stories/hooks/delete-mutation";
import { StoryPriority } from "@/modules/stories/types";
import { PriorityIcon } from "./priority-icon";
import { useSprints } from "@/lib/hooks/sprints";
import { useTeams } from "@/lib/hooks/teams";
import { useStatuses } from "@/lib/hooks/statuses";
import { useObjectives } from "@/lib/hooks/objectives";

export const StoriesToolbar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const { selectedStories, setSelectedStories } = useBoard();

  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useStatuses();
  const { data: sprints = [] } = useSprints();
  const { data: objectives = [] } = useObjectives();

  const { mutateAsync: bulkDeleteMutate } = useBulkDeleteStoryMutation();

  const handleBulkDelete = async () => {
    bulkDeleteMutate(selectedStories);
    setSelectedStories([]);
    setIsOpen(false);
  };

  const priorities: StoryPriority[] = [
    "No Priority",
    "Low",
    "Medium",
    "High",
    "Urgent",
  ];
  return (
    <>
      <Flex
        align="center"
        className="fixed bottom-8 left-1/2 right-1/2 z-50 w-max -translate-x-1/2 rounded-[0.55rem] border border-gray-100 bg-white/60 px-2.5 py-2 shadow-lg shadow-dark/10 backdrop-blur dark:border-dark-50 dark:bg-dark-300/70 dark:shadow-dark/20"
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
          onClick={() => {
            setIsSprintsOpen(true);
          }}
          leftIcon={<SprintsIcon className="h-[1.15rem] w-auto" />}
          variant="outline"
        >
          Add to sprint
        </Button>
        <Button
          color="tertiary"
          onClick={() => {
            setIsObjectivesOpen(true);
          }}
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
                    value={name}
                    onSelect={() => {
                      setIsSprintsOpen(false);
                    }}
                    className="justify-between py-2.5"
                    key={id}
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
