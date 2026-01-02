"use client";
import { Button, Flex, Text, Tooltip, Dialog, DatePicker } from "ui";
import {
  ArchiveIcon,
  AssigneeIcon,
  CalendarIcon,
  CloseIcon,
  DeleteIcon,
  ObjectiveIcon,
  SprintsIcon,
  UndoIcon,
} from "icons";
import { useState } from "react";
import { useParams, usePathname } from "next/navigation";
import { formatISO } from "date-fns";
import { useBulkDeleteStoryMutation } from "@/modules/stories/hooks/delete-mutation";
import { useBulkArchiveStoryMutation } from "@/modules/stories/hooks/archive-mutation";
import { useBulkUnarchiveStoryMutation } from "@/modules/stories/hooks/unarchive-mutation";
import { useBulkRestoreStoryMutation } from "@/modules/stories/hooks/restore-mutation";
import { useTerminology } from "@/hooks";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useBulkUpdateStoriesMutation } from "@/modules/stories/hooks/update-mutation";
import type { DetailedStory } from "@/modules/story/types";
import { useBoard } from "./board-context";
import { StoryStatusIcon } from "./story-status-icon";
import { PriorityIcon } from "./priority-icon";
import { SprintsMenu } from "./story/sprints-menu";
import { ObjectivesMenu } from "./story/objectives-menu";
import { StatusesMenu } from "./story/statuses-menu";
import { PrioritiesMenu } from "./story/priorities-menu";
import { AssigneesMenu } from "./story/assignees-menu";

export const StoriesToolbar = () => {
  const pathname = usePathname();
  const { teamId } = useParams<{ teamId: string }>();
  const { getTermDisplay } = useTerminology();
  const [isOpen, setIsOpen] = useState(false);
  const [isArchiveDialogOpen, setIsArchiveDialogOpen] = useState(false);
  const [isUnarchiveDialogOpen, setIsUnarchiveDialogOpen] = useState(false);
  const [isRestoreDialogOpen, setIsRestoreDialogOpen] = useState(false);
  const { selectedStories, setSelectedStories } = useBoard();
  const { data: teams = [] } = useTeams();
  let finalTeamId = teamId;
  if (teams.length === 1) {
    finalTeamId = teams[0].id;
  }
  const isOnDeletedStoriesPage = pathname.includes("/deleted");
  const isOnArchivePage = pathname.includes("/archived");

  const { mutate: bulkDeleteMutate, isPending } = useBulkDeleteStoryMutation();
  const { mutate: bulkArchiveMutate } = useBulkArchiveStoryMutation();
  const { mutate: bulkUnarchiveMutate } = useBulkUnarchiveStoryMutation();
  const { mutate: bulkRestoreMutate } = useBulkRestoreStoryMutation();
  const { mutate: bulkUpdateMutate } = useBulkUpdateStoriesMutation();

  const handleBulkDelete = () => {
    bulkDeleteMutate({
      storyIds: selectedStories,
      hardDelete: isOnDeletedStoriesPage || isOnArchivePage,
    });
    setSelectedStories([]);
    setIsOpen(false);
  };

  const handleBulkUpdate = (updates: Partial<DetailedStory>) => {
    bulkUpdateMutate({
      storyIds: selectedStories,
      payload: updates,
    });
  };

  const handleBulkArchive = () => {
    bulkArchiveMutate(selectedStories);
    setSelectedStories([]);
    setIsArchiveDialogOpen(false);
  };

  const handleBulkUnarchive = () => {
    bulkUnarchiveMutate(selectedStories);
    setSelectedStories([]);
    setIsUnarchiveDialogOpen(false);
  };

  const handleBulkRestore = () => {
    bulkRestoreMutate(selectedStories);
    setSelectedStories([]);
    setIsRestoreDialogOpen(false);
  };

  return (
    <>
      <Flex
        align="center"
        className="border-border bg-surface-elevated/90 shadow-shadow fixed right-1/2 bottom-8 left-1/2 z-50 w-max -translate-x-1/2 rounded-2xl border-[0.5px] px-2.5 py-2 shadow-lg backdrop-blur"
        gap={1}
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
              size="sm"
              variant="outline"
            >
              <span className="sr-only">Clear</span>
            </Button>
          </Tooltip>
          {selectedStories.length} selected
        </Text>
        {finalTeamId && !isOnDeletedStoriesPage && !isOnArchivePage ? (
          <>
            <SprintsMenu>
              <SprintsMenu.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<SprintsIcon className="h-[1.15rem]" />}
                  variant="naked"
                >
                  {getTermDisplay("sprintTerm", { capitalize: true })}
                </Button>
              </SprintsMenu.Trigger>
              <SprintsMenu.Items
                setSprintId={(sprintId) => {
                  handleBulkUpdate({ sprintId });
                }}
                teamId={finalTeamId}
              />
            </SprintsMenu>
            <ObjectivesMenu>
              <ObjectivesMenu.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<ObjectiveIcon className="h-[1.15rem]" />}
                  variant="naked"
                >
                  {getTermDisplay("objectiveTerm", { capitalize: true })}
                </Button>
              </ObjectivesMenu.Trigger>
              <ObjectivesMenu.Items
                setObjectiveId={(objectiveId) => {
                  handleBulkUpdate({ objectiveId });
                }}
                teamId={finalTeamId}
              />
            </ObjectivesMenu>
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={
                    <StoryStatusIcon
                      category="started"
                      className="text-warning dark:text-warning h-[1.15rem]"
                    />
                  }
                  variant="naked"
                >
                  Status
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items
                setStatusId={(statusId) => {
                  handleBulkUpdate({ statusId });
                }}
                teamId={finalTeamId}
              />
            </StatusesMenu>
          </>
        ) : null}

        {!isOnArchivePage && (
          <PrioritiesMenu>
            <PrioritiesMenu.Trigger>
              <Button
                color="tertiary"
                leftIcon={
                  <PriorityIcon
                    className="text-success dark:text-success h-[1.15rem]"
                    priority="High"
                  />
                }
                variant="naked"
              >
                Priority
              </Button>
            </PrioritiesMenu.Trigger>
            <PrioritiesMenu.Items
              setPriority={(priority) => {
                handleBulkUpdate({ priority });
              }}
            />
          </PrioritiesMenu>
        )}

        {!isOnArchivePage && (
          <DatePicker>
            <DatePicker.Trigger>
              <Button
                color="tertiary"
                leftIcon={
                  <CalendarIcon className="text-primary dark:text-primary h-[1.15rem]" />
                }
                variant="naked"
              >
                Deadline
              </Button>
            </DatePicker.Trigger>
            <DatePicker.Calendar
              onDayClick={(day) => {
                handleBulkUpdate({ endDate: formatISO(day) });
              }}
            />
          </DatePicker>
        )}

        {finalTeamId && !isOnDeletedStoriesPage && !isOnArchivePage ? (
          <AssigneesMenu>
            <AssigneesMenu.Trigger>
              <Button
                color="tertiary"
                leftIcon={<AssigneeIcon className="h-[1.15rem]" />}
                variant="naked"
              >
                Assign to...
              </Button>
            </AssigneesMenu.Trigger>
            <AssigneesMenu.Items
              onAssigneeSelected={(assigneeId) => {
                handleBulkUpdate({ assigneeId });
              }}
              teamId={finalTeamId}
            />
          </AssigneesMenu>
        ) : null}
        {!isOnDeletedStoriesPage && !isOnArchivePage && (
          <Button
            color="tertiary"
            leftIcon={<ArchiveIcon className="h-[1.15rem]" />}
            onClick={() => {
              setIsArchiveDialogOpen(true);
            }}
            variant="naked"
          >
            Archive
          </Button>
        )}

        {isOnArchivePage ? (
          <Button
            color="tertiary"
            leftIcon={<ArchiveIcon className="h-[1.15rem]" />}
            onClick={() => {
              setIsUnarchiveDialogOpen(true);
            }}
            variant="naked"
          >
            Unarchive
          </Button>
        ) : null}

        {isOnDeletedStoriesPage ? (
          <Button
            color="tertiary"
            leftIcon={<UndoIcon className="h-[1.15rem]" />}
            onClick={() => {
              setIsRestoreDialogOpen(true);
            }}
            variant="naked"
          >
            Restore
          </Button>
        ) : null}

        <Button
          leftIcon={<DeleteIcon className="h-[1.15rem] text-white" />}
          onClick={() => {
            setIsOpen(true);
          }}
        >
          {isOnDeletedStoriesPage || isOnArchivePage
            ? "Delete forever"
            : "Delete"}
        </Button>
      </Flex>

      {/* Archive stories dialog */}
      <Dialog onOpenChange={setIsArchiveDialogOpen} open={isArchiveDialogOpen}>
        <Dialog.Content>
          <Dialog.Header className="px-6 pt-6">
            <Dialog.Title className="text-lg">
              Archive {selectedStories.length} stories?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              These stories will be moved to the archive and can be unarchived
              later. They won&apos;t appear in your active story lists.
            </Text>
            <Flex align="center" className="mt-4" gap={2} justify="end">
              <Button
                color="tertiary"
                onClick={() => {
                  setIsArchiveDialogOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                leftIcon={
                  <ArchiveIcon className="text-white dark:text-white" />
                }
                onClick={handleBulkArchive}
              >
                Archive
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>

      {/* Unarchive stories dialog */}
      <Dialog
        onOpenChange={setIsUnarchiveDialogOpen}
        open={isUnarchiveDialogOpen}
      >
        <Dialog.Content>
          <Dialog.Header className="px-6 pt-6">
            <Dialog.Title className="text-lg">
              Unarchive {selectedStories.length} stories?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              These stories will be restored to your active story list and can
              be assigned to sprints and team members again.
            </Text>
            <Flex align="center" className="mt-4" gap={2} justify="end">
              <Button
                color="tertiary"
                onClick={() => {
                  setIsUnarchiveDialogOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                leftIcon={
                  <ArchiveIcon className="text-white dark:text-white" />
                }
                onClick={handleBulkUnarchive}
              >
                Unarchive
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>

      {/* Restore stories dialog */}
      <Dialog onOpenChange={setIsRestoreDialogOpen} open={isRestoreDialogOpen}>
        <Dialog.Content>
          <Dialog.Header className="px-6 pt-6">
            <Dialog.Title className="text-lg">
              Restore {selectedStories.length} stories?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              These stories will be restored to your active story list and can
              be assigned to sprints and team members again.
            </Text>
            <Flex align="center" className="mt-4" gap={2} justify="end">
              <Button
                color="tertiary"
                onClick={() => {
                  setIsRestoreDialogOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                leftIcon={<UndoIcon className="text-white dark:text-white" />}
                onClick={handleBulkRestore}
              >
                Restore
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>

      {/* Delete stories dialog */}
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content>
          <Dialog.Header className="px-6 pt-6">
            <Dialog.Title className="text-lg">
              Delete {selectedStories.length} stories
              {isOnDeletedStoriesPage ? " forever?" : "?"}
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              {isOnDeletedStoriesPage
                ? "This is an irreversible action. The stories will be permanently deleted. You can't restore them."
                : "These stories will be moved to the recycle bin and will be permanently deleted after 30 days. You can restore them at any time before that."}
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
                  <DeleteIcon className="text-white" />
                }
                loading={isPending}
                loadingText="Deleting..."
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
