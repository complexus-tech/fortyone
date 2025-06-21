"use client";
import { Button, Flex, Text, Tooltip, Dialog, DatePicker } from "ui";
import {
  AssigneeIcon,
  CalendarIcon,
  CloseIcon,
  DeleteIcon,
  ObjectiveIcon,
  SprintsIcon,
} from "icons";
import { useState } from "react";
import { useParams } from "next/navigation";
import { formatISO } from "date-fns";
import { useBulkDeleteStoryMutation } from "@/modules/stories/hooks/delete-mutation";
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
  const { teamId } = useParams<{ teamId: string }>();
  const { getTermDisplay } = useTerminology();
  const [isOpen, setIsOpen] = useState(false);
  const { selectedStories, setSelectedStories } = useBoard();
  const { data: teams = [] } = useTeams();
  let finalTeamId = teamId;
  if (teams.length === 1) {
    finalTeamId = teams[0].id;
  }

  const { mutate: bulkDeleteMutate, isPending } = useBulkDeleteStoryMutation();
  const { mutate: bulkUpdateMutate } = useBulkUpdateStoriesMutation();

  const handleBulkDelete = () => {
    bulkDeleteMutate(selectedStories);
    setSelectedStories([]);
    setIsOpen(false);
  };

  const handleBulkUpdate = (updates: Partial<DetailedStory>) => {
    bulkUpdateMutate({
      storyIds: selectedStories,
      payload: updates,
    });
  };

  return (
    <>
      <Flex
        align="center"
        className="fixed bottom-8 left-1/2 right-1/2 z-50 w-max -translate-x-1/2 rounded-2xl border-[0.5px] border-gray-200/70 bg-gray-50/70 px-2.5 py-2 shadow-lg shadow-gray-200 backdrop-blur dark:border-dark-50 dark:bg-dark-200/70 dark:shadow-dark/20"
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
        {finalTeamId ? (
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
                      className="h-[1.15rem] text-warning dark:text-warning"
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

        <PrioritiesMenu>
          <PrioritiesMenu.Trigger>
            <Button
              color="tertiary"
              leftIcon={
                <PriorityIcon
                  className="h-[1.15rem] text-success dark:text-success"
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

        <DatePicker>
          <DatePicker.Trigger>
            <Button
              color="tertiary"
              leftIcon={
                <CalendarIcon className="h-[1.15rem] text-primary dark:text-primary" />
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

        {finalTeamId ? (
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
