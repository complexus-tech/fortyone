"use client";
import Link from "next/link";
import {
  DatePicker,
  Flex,
  Text,
  Tooltip,
  Avatar,
  Checkbox,
  Box,
  Button,
  Badge,
} from "ui";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Story as StoryProps } from "@/modules/stories/types";
import { RowWrapper } from "../row-wrapper";
import { StoryStatusIcon } from "../story-status-icon";
import { PriorityIcon } from "../priority-icon";
import { useBoard } from "../board-context";
import { AssigneesMenu } from "./assignees-menu";
import { StoryContextMenu } from "./context-menu";
import { DragHandle } from "./drag-handle";
import { Labels } from "./labels";
import { PrioritiesMenu } from "./priorities-menu";
import { StatusesMenu } from "./statuses-menu";
import { slugify } from "@/utils";
import { CalendarIcon, ObjectiveIcon, SprintsIcon, CloseIcon } from "icons";
import {
  format,
  addDays,
  differenceInDays,
  isTomorrow,
  formatISO,
} from "date-fns";
import { StateCategory } from "@/types/states";
import { DetailedStory } from "@/modules/story/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useStatuses } from "@/lib/hooks/statuses";
import { useSprints } from "@/lib/hooks/sprints";
import { SprintsMenu } from "@/components/ui";
import { useMembers } from "@/lib/hooks/members";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { ObjectivesMenu } from "./objectives-menu";

export const StoryRow = ({ story }: { story: StoryProps }) => {
  const {
    id,
    sequenceId,
    title,
    statusId,
    endDate,
    createdAt,
    updatedAt,
    teamId,
    objectiveId,
    sprintId,
    priority = "No Priority",
  } = story;
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useStatuses();
  const { data: sprints = [] } = useSprints();
  const { data: objectives = [] } = useObjectives();
  const { data: members = [] } = useMembers();
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id,
  });
  const { selectedStories, setSelectedStories, isColumnVisible } = useBoard();

  const status =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  const teamCode = teams.find((team) => team.id === teamId)?.code;
  const selectedObjective = objectives.find(
    (objective) => objective.id === objectiveId,
  );
  const selectedSprint = sprints.find((sprint) => sprint.id === sprintId);
  const selectedAssignee = members.find(
    (member) => member.id === story?.assigneeId,
  );
  const completedOrCancelled = (category?: StateCategory) => {
    return ["completed", "cancelled", "paused"].includes(category || "");
  };

  const getDueDateMessage = (date: Date) => {
    if (date < new Date()) {
      const daysOverdue = differenceInDays(new Date(), date);
      if (daysOverdue === 0) {
        return (
          <>
            <Text fontSize="md">The story is due today</Text>
            <Text color="muted" fontSize="md">
              Zero days overdue
            </Text>
          </>
        );
      }
      if (daysOverdue === 1) {
        return (
          <>
            <Text fontSize="md">This was due on yesterday</Text>
            <Text color="muted" fontSize="md">
              One day overdue
            </Text>
          </>
        );
      }
      return (
        <>
          <Text fontSize="md">
            This was due on {format(date, "MMM d, yyyy")}
          </Text>
          <Text color="muted" fontSize="md">
            {differenceInDays(new Date(), date)} days overdue
          </Text>
        </>
      );
    }
    if (date <= addDays(new Date(), 7) && date >= new Date()) {
      return (
        <>
          <Text fontSize="md">Due on {format(date, "MMM d, yyyy")}</Text>
          <Text fontSize="md" color="muted">
            {isTomorrow(date) ? (
              "Due tomorrow"
            ) : (
              <>Due in {differenceInDays(date, new Date()) + 1} days</>
            )}
          </Text>
        </>
      );
    }
    return (
      <>
        <Text fontSize="md">Due on {format(date, "MMM d, yyyy")}</Text>
        <Text color="muted" fontSize="md">
          {isTomorrow(date) ? (
            "Tomorrow"
          ) : (
            <>Due in {differenceInDays(date, new Date()) + 1} days</>
          )}
        </Text>
      </>
    );
  };

  const sprintTooltip = () => {
    const isCompleted = new Date(selectedSprint?.endDate!!) < new Date();
    const inProgress =
      new Date(selectedSprint?.startDate!!) < new Date() &&
      new Date(selectedSprint?.endDate!!) > new Date();
    const daysLeft = differenceInDays(
      new Date(selectedSprint?.endDate!!),
      new Date(),
    );
    const isPanned = new Date(selectedSprint?.startDate!!) > new Date();

    const getBadgeColor = () => {
      if (isCompleted || isPanned) {
        return "tertiary";
      }
      if (inProgress && daysLeft < 5) {
        return "primary";
      }
      if (inProgress && daysLeft < 8) {
        return "warning";
      }
      return "info";
    };

    const getBadgeText = () => {
      if (isCompleted) {
        return "Completed";
      }
      if (isPanned) {
        return "Planned";
      }
      if (inProgress && daysLeft < 5) {
        return `${daysLeft} days left`;
      }
      if (inProgress && daysLeft < 8) {
        return "Ending in a week";
      }
      return "In progress";
    };

    return (
      <Box>
        <Text fontSize="md" className="flex items-center gap-1">
          <SprintsIcon className="h-5 w-auto shrink-0" />
          {selectedSprint?.name}
        </Text>
        <Flex align="center" gap={6} className="mb-3 mt-4" justify="between">
          <Text fontSize="md" className="flex items-center gap-1">
            <CalendarIcon
              className={cn("h-5 w-auto", {
                "text-primary dark:text-primary": getBadgeColor() === "primary",
                "text-warning dark:text-warning": getBadgeColor() === "warning",
                "text-info dark:text-info": getBadgeColor() === "info",
              })}
            />{" "}
            {format(new Date(selectedSprint?.startDate!!), "MMM dd")} -{" "}
            {format(new Date(selectedSprint?.endDate!!), "MMM dd")}
          </Text>
          <Badge
            className="h-7 border-opacity-50 bg-opacity-40 text-[0.95rem] font-medium capitalize"
            color={getBadgeColor()}
          >
            {getBadgeText()}
          </Badge>
        </Flex>
        {selectedSprint?.goal && (
          <>
            <Text fontSize="md">Sprint Goal:</Text>
            <Text color="muted" className="mt-1 line-clamp-4" fontSize="md">
              {selectedSprint?.goal}
            </Text>
          </>
        )}
      </Box>
    );
  };

  const { mutateAsync } = useUpdateStoryMutation();

  const handleUpdate = async (data: Partial<DetailedStory>) => {
    mutateAsync({
      storyId: id,
      payload: data,
    });
  };

  return (
    <div ref={setNodeRef}>
      <StoryContextMenu story={story}>
        <RowWrapper
          className={cn("gap-4", {
            "bg-gray-50 opacity-70 dark:bg-dark-50/40 dark:opacity-50":
              isDragging,
          })}
        >
          <Flex align="center" className="relative shrink select-none" gap={2}>
            <DragHandle {...listeners} {...attributes} />
            <Checkbox
              checked={selectedStories.includes(id)}
              className="absolute -left-[1.6rem] rounded-[0.35rem]"
              onCheckedChange={(checked) => {
                setSelectedStories(
                  checked
                    ? [...selectedStories, id]
                    : selectedStories.filter((storyId) => storyId !== id),
                );
              }}
            />
            {isColumnVisible("ID") && (
              <Tooltip title={`Story ID: ${teamCode}-${sequenceId}`}>
                <Text
                  className="min-w-[6ch] truncate text-[0.99rem]"
                  color="muted"
                >
                  {teamCode}-{sequenceId}
                </Text>
              </Tooltip>
            )}
            <Link href={`/story/${id}/${slugify(title)}`}>
              <Text
                fontWeight="medium"
                className="line-clamp-1 hover:opacity-90"
              >
                {title}
              </Text>
            </Link>
          </Flex>
          <Flex align="center" className="shrink-0" gap={3}>
            {isColumnVisible("Status") && (
              <StatusesMenu>
                <StatusesMenu.Trigger>
                  <button className="flex items-center gap-1" type="button">
                    <StoryStatusIcon statusId={statusId} /> {status?.name}
                  </button>
                </StatusesMenu.Trigger>
                <StatusesMenu.Items
                  statusId={statusId}
                  setStatusId={(id) => {
                    handleUpdate({ statusId: id });
                  }}
                />
              </StatusesMenu>
            )}
            {isColumnVisible("Priority") && (
              <PrioritiesMenu>
                <PrioritiesMenu.Trigger>
                  <button
                    className="flex select-none items-center gap-1"
                    type="button"
                  >
                    <PriorityIcon priority={priority} />
                    {priority}
                  </button>
                </PrioritiesMenu.Trigger>
                <PrioritiesMenu.Items
                  priority={priority}
                  setPriority={(p) => {
                    handleUpdate({ priority: p });
                  }}
                />
              </PrioritiesMenu>
            )}
            {isColumnVisible("Objective") && selectedObjective && (
              <ObjectivesMenu>
                <Tooltip
                  className="max-w-80 py-3"
                  title={
                    <Flex align="start" gap={2}>
                      <ObjectiveIcon className="relative top-[3px] h-4 shrink-0" />
                      <Box>
                        <Text fontSize="md" className="mb-1.5">
                          {selectedObjective?.name}
                        </Text>
                        <Text
                          color="muted"
                          className="line-clamp-4"
                          fontSize="md"
                        >
                          {selectedObjective?.description}
                        </Text>
                      </Box>
                    </Flex>
                  }
                >
                  <span>
                    <ObjectivesMenu.Trigger>
                      <Button
                        color="tertiary"
                        className="gap-1 px-2"
                        size="xs"
                        rounded="xl"
                        type="button"
                        rightIcon={
                          objectiveId && (
                            <span
                              tabIndex={0}
                              role="button"
                              aria-label="Remove objective"
                              className="flex aspect-square items-center gap-1"
                              onClick={() => {
                                handleUpdate({ objectiveId: null });
                              }}
                            >
                              <CloseIcon className="h-4" strokeWidth={3} />
                              <span className="sr-only">Remove objective</span>
                            </span>
                          )
                        }
                      >
                        <ObjectiveIcon className="h-4" />
                        <span className="inline-block max-w-36 truncate">
                          {selectedObjective?.name}
                        </span>
                      </Button>
                    </ObjectivesMenu.Trigger>
                  </span>
                </Tooltip>
                <ObjectivesMenu.Items
                  objectiveId={objectiveId ?? undefined}
                  setObjectiveId={(objectiveId) => {
                    handleUpdate({ objectiveId });
                  }}
                />
              </ObjectivesMenu>
            )}

            {isColumnVisible("Sprint") && selectedSprint && (
              <SprintsMenu>
                <Tooltip
                  className="pointer-events-none max-w-96 py-3"
                  title={sprintTooltip()}
                >
                  <span>
                    <SprintsMenu.Trigger>
                      <Button
                        color="tertiary"
                        className="gap-1 pl-1.5 pr-2"
                        size="xs"
                        rounded="xl"
                        type="button"
                      >
                        <SprintsIcon className="relative -top-[0.3px] h-[1.1rem]" />
                        <span className="inline-block max-w-36 truncate">
                          {selectedSprint?.name}
                        </span>
                      </Button>
                    </SprintsMenu.Trigger>
                  </span>
                </Tooltip>
                <SprintsMenu.Items
                  sprintId={sprintId ?? undefined}
                  setSprintId={(sprintId) => {
                    handleUpdate({ sprintId });
                  }}
                />
              </SprintsMenu>
            )}
            {isColumnVisible("Labels") && <Labels />}
            {isColumnVisible("Due date") &&
              endDate &&
              !completedOrCancelled(status?.category) && (
                <DatePicker>
                  <Tooltip
                    className="py-3"
                    title={
                      <Flex align="start" gap={2}>
                        <CalendarIcon
                          className={cn("relative top-[2.5px] h-5 w-auto", {
                            "text-primary dark:text-primary":
                              new Date(endDate) < new Date(),
                            "text-warning dark:text-warning":
                              new Date(endDate) <= addDays(new Date(), 7) &&
                              new Date(endDate) >= new Date(),
                          })}
                        />
                        <Box>{getDueDateMessage(new Date(endDate))}</Box>
                      </Flex>
                    }
                  >
                    <span>
                      <DatePicker.Trigger>
                        <Button
                          color="tertiary"
                          className={cn("px-2", {
                            "text-primary dark:text-primary":
                              new Date(endDate) < new Date(),
                            "text-warning dark:text-warning":
                              new Date(endDate) <= addDays(new Date(), 7) &&
                              new Date(endDate) >= new Date(),
                          })}
                          size="xs"
                          rounded="xl"
                          type="button"
                        >
                          <CalendarIcon
                            className={cn("h-4", {
                              "text-primary dark:text-primary":
                                new Date(endDate) < new Date(),
                              "text-warning dark:text-warning":
                                new Date(endDate) <= addDays(new Date(), 7) &&
                                new Date(endDate) >= new Date(),
                            })}
                            strokeWidth={3}
                          />
                          {format(new Date(endDate), "MMM d")}
                        </Button>
                      </DatePicker.Trigger>
                    </span>
                  </Tooltip>
                  <DatePicker.Calendar
                    selected={new Date(endDate)}
                    onDayClick={(day) => {
                      handleUpdate({ endDate: formatISO(day) });
                    }}
                  />
                </DatePicker>
              )}
            {isColumnVisible("Created") && (
              <Tooltip
                title={`Created on ${format(new Date(createdAt), "MMM dd, yyyy HH:mm")}`}
              >
                <span className="cursor-default">
                  <Text as="span" color="muted">
                    {format(new Date(createdAt), "MMM dd")}
                  </Text>
                </span>
              </Tooltip>
            )}
            {isColumnVisible("Updated") && (
              <Tooltip
                title={`Last updated on ${format(new Date(updatedAt), "MMM dd, yyyy HH:mm")}`}
              >
                <span className="cursor-default">
                  <Text as="span" color="muted">
                    {format(new Date(updatedAt), "MMM dd")}
                  </Text>
                </span>
              </Tooltip>
            )}
            {isColumnVisible("Assignee") && (
              <AssigneesMenu>
                <Tooltip
                  className="mr-2 py-2.5"
                  title={
                    selectedAssignee ? (
                      <Box>
                        <Flex gap={2}>
                          <Avatar
                            name={selectedAssignee?.fullName}
                            src={selectedAssignee?.avatarUrl}
                            className="mt-0.5"
                          />
                          <Box>
                            <Link
                              href={`/profile/${selectedAssignee?.id}`}
                              className="mb-2 flex gap-1"
                            >
                              <Text fontWeight="medium" fontSize="md">
                                {selectedAssignee?.fullName}
                              </Text>
                              <Text color="muted" fontSize="md">
                                ({selectedAssignee?.username})
                              </Text>
                            </Link>
                            <Button
                              size="xs"
                              color="tertiary"
                              className="mb-0.5 ml-px px-2"
                              href={`/profile/${selectedAssignee?.id}`}
                            >
                              Go to profile
                            </Button>
                          </Box>
                        </Flex>
                      </Box>
                    ) : null
                  }
                >
                  <span>
                    <AssigneesMenu.Trigger>
                      <button className="flex" type="button">
                        <Avatar
                          color="tertiary"
                          name={selectedAssignee?.fullName}
                          size="sm"
                          src={selectedAssignee?.avatarUrl}
                        />
                      </button>
                    </AssigneesMenu.Trigger>
                  </span>
                </Tooltip>
                <AssigneesMenu.Items
                  assigneeId={selectedAssignee?.id}
                  onAssigneeSelected={(assigneeId) => {
                    handleUpdate({ assigneeId });
                  }}
                />
              </AssigneesMenu>
            )}
          </Flex>
        </RowWrapper>
      </StoryContextMenu>
    </div>
  );
};
