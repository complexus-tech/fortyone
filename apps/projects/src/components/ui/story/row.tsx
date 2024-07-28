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
import { useStore } from "@/hooks/store";
import {
  CalendarIcon,
  EpicsIcon,
  LinkIcon,
  ObjectiveIcon,
  SprintsIcon,
} from "icons";
import { format, addDays, differenceInDays, isTomorrow } from "date-fns";
import { StateCategory } from "@/types/states";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { DetailedStory } from "@/modules/story/types";
import { toast } from "sonner";

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
    epicId,
    priority = "No Priority",
  } = story;
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id,
  });
  const { selectedStories, setSelectedStories, isColumnVisible } = useBoard();
  const { states, teams, objectives, sprints } = useStore();
  const status = states.find((state) => state.id === statusId) || states.at(0);
  const teamCode = teams.find((team) => team.id === teamId)?.code;
  const selectedObjective = objectives.find(
    (objective) => objective.id === objectiveId,
  );
  const selectedSprint = sprints.find((sprint) => sprint.id === sprintId);

  const completedOrCancelled = (category?: StateCategory) => {
    return ["completed", "cancelled", "paused"].includes(category || "");
  };

  const getDueDateMessage = (date: Date) => {
    if (date < new Date()) {
      return (
        <>
          <Text fontSize="md">
            This was overdue on {format(date, "MMMM dd")}
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
          <Text fontSize="md">Due on {format(date, "MMMM dd")}</Text>
          <Text fontSize="md" color="muted">
            {isTomorrow(date) ? (
              "Tomorrow"
            ) : (
              <>Due in {differenceInDays(date, new Date())} days</>
            )}
          </Text>
        </>
      );
    }
    return (
      <>
        <Text fontSize="md">Due on {format(date, "MMMM dd")}</Text>
        <Text color="muted" fontSize="md">
          {isTomorrow(date) ? (
            "Tomorrow"
          ) : (
            <>Due in {differenceInDays(date, new Date())} days</>
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
      <Flex align="start" gap={2}>
        <SprintsIcon className="relative top-[6px] h-5 w-auto shrink-0" />
        <Box>
          <Flex align="center" justify="between">
            <Text fontSize="md">{selectedSprint?.name}</Text>

            <Button
              color="tertiary"
              rounded="full"
              size="sm"
              title="Open sprint"
            >
              <LinkIcon className="h-4 w-auto -rotate-45" />
              <span className="sr-only">Open sprint</span>
            </Button>
          </Flex>
          <Flex align="center" gap={6} className="mb-2 mt-3" justify="between">
            <Text fontSize="md" className="flex items-center gap-1">
              <CalendarIcon
                className={cn("h-5 w-auto", {
                  "text-primary dark:text-primary":
                    getBadgeColor() === "primary",
                  "text-warning dark:text-warning":
                    getBadgeColor() === "warning",
                  "text-info dark:text-info": getBadgeColor() === "info",
                })}
              />{" "}
              {format(new Date(selectedSprint?.startDate!!), "MMM dd")} -{" "}
              {format(new Date(selectedSprint?.endDate!!), "MMM dd")}
            </Text>
            <Badge
              className="border-opacity-50 bg-opacity-40 text-xs font-semibold uppercase"
              rounded="full"
              color={getBadgeColor()}
            >
              {getBadgeText()}
            </Badge>
          </Flex>
          {selectedSprint?.goal && (
            <>
              <Text fontSize="md">Sprint Goal:</Text>
              <Text color="muted" className="line-clamp-4" fontSize="md">
                {selectedSprint?.goal}
              </Text>
            </>
          )}
        </Box>
      </Flex>
    );
  };
  const handleUpdate = async (data: Partial<DetailedStory>) => {
    try {
      await updateStoryAction(id, data);
    } catch (error) {
      toast.error("Error updating story", {
        description: "Failed to update story, reverted changes.",
      });
    }
  };

  return (
    <div ref={setNodeRef}>
      <StoryContextMenu>
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
              <Text className="line-clamp-1 hover:opacity-90">{title}</Text>
            </Link>
          </Flex>
          <Flex align="center" className="shrink-0" gap={2}>
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
              <Tooltip
                className="max-w-80 py-3"
                title={
                  <Flex align="start" gap={2}>
                    <ObjectiveIcon className="relative top-[2.5px] h-5 w-auto shrink-0" />
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
                <Button
                  color="tertiary"
                  className="pl-1.5 pr-2"
                  size="xs"
                  rounded="xl"
                  type="button"
                >
                  <ObjectiveIcon className="h-5 w-auto" />
                  <span className="inline-block max-w-36 truncate">
                    {selectedObjective?.name}
                  </span>
                </Button>
              </Tooltip>
            )}

            {isColumnVisible("Sprint") && selectedSprint && (
              <Tooltip className="max-w-96 py-3" title={sprintTooltip()}>
                <Button
                  color="tertiary"
                  className="pl-1.5 pr-2"
                  size="xs"
                  rounded="xl"
                  type="button"
                >
                  <SprintsIcon className="h-5 w-auto" />
                  <span className="inline-block max-w-36 truncate">
                    {selectedSprint?.name}
                  </span>
                </Button>
              </Tooltip>
            )}

            <Button
              color="tertiary"
              className="px-2"
              size="xs"
              rounded="xl"
              type="button"
            >
              <EpicsIcon className="h-5 w-auto" /> Objective
            </Button>
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
                      <DatePicker.Trigger asChild>
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
                            className="h-4 w-auto"
                            strokeWidth={2.5}
                          />
                          {format(new Date(endDate), "MMM dd")}
                        </Button>
                      </DatePicker.Trigger>
                    </span>
                  </Tooltip>
                  <DatePicker.Calendar
                    selected={new Date(endDate)}
                    onDayClick={(day) => {
                      handleUpdate({ endDate: day.toISOString() });
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
                  title={
                    <Flex align="center" gap={2}>
                      <Avatar
                        name="Joseph Mukorivo"
                        src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                      />
                      <Box>
                        <Text fontSize="md" fontWeight="medium">
                          Joseph Mukorivo
                        </Text>
                        <Text fontSize="md" color="muted">
                          @josemukorivo
                        </Text>
                      </Box>
                    </Flex>
                  }
                >
                  <span>
                    <AssigneesMenu.Trigger>
                      <button className="flex" type="button">
                        <Avatar
                          color="tertiary"
                          // name="Joseph Mukorivo"
                          size="xs"
                          // src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                        />
                      </button>
                    </AssigneesMenu.Trigger>
                  </span>
                </Tooltip>
                <AssigneesMenu.Items />
              </AssigneesMenu>
            )}
          </Flex>
        </RowWrapper>
      </StoryContextMenu>
    </div>
  );
};
