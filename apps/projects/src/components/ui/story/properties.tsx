import {
  Box,
  Flex,
  Button,
  Text,
  DatePicker,
  Tooltip,
  Badge,
  Divider,
} from "ui";
import {
  ArrowRight2Icon,
  CalendarIcon,
  EstimateIcon,
  ObjectiveIcon,
  SprintsIcon,
  SubStoryIcon,
} from "icons";
import { cn } from "lib";
import { format, addDays, formatISO } from "date-fns";
import Link from "next/link";
import { useRef, useState } from "react";
import { ObjectivesMenu } from "@/components/ui/story/objectives-menu";
import { SprintsMenu } from "@/components/ui/story/sprints-menu";
import { EstimateMenu } from "@/components/ui/story/estimate-menu";
import { Labels } from "@/components/ui/story/labels";
import { sprintTooltip } from "@/components/ui/story/sprint-tooltip";
import { getDueDateMessage } from "@/components/ui/story/due-date-tooltip";
import { StoryStatusIcon } from "@/components/ui/story-status-icon";
import { PriorityIcon } from "@/components/ui/priority-icon";
import { StatusesMenu } from "@/components/ui/story/statuses-menu";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import type { Story } from "@/modules/stories/types";
import { useBoard } from "@/components/ui/board-context";
import type { StateCategory } from "@/types/states";
import { useTerminology, useUserRole, useWorkspacePath } from "@/hooks";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { hexToRgba, slugify } from "@/utils";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { formatEstimate } from "@/lib/estimate";
import { RowWrapper } from "../row-wrapper";

type StoryPropertiesProps = Story & {
  handleUpdate: (data: Partial<Story>) => void;
  asKanban?: boolean;
  teamCode?: string;
  isExpanded?: boolean;
  setIsExpanded?: (isExpanded: boolean) => void;
};

const completedOrCancelled = (category?: StateCategory) => {
  return ["completed", "cancelled", "paused"].includes(category || "");
};

export const StoryProperties = ({
  handleUpdate,
  statusId,
  priority,
  estimateValue,
  estimateScheme,
  objective,
  objectiveId,
  sprint,
  sprintId,
  id,
  teamId,
  endDate,
  createdAt,
  updatedAt,
  labels: storyLabels,
  asKanban,
  subStories,
  teamCode,
  isExpanded,
  setIsExpanded,
}: StoryPropertiesProps) => {
  const { getTermDisplay } = useTerminology();
  const { isColumnVisible } = useBoard();
  const { withWorkspace } = useWorkspacePath();
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const [showChildrenDialog, setShowChildrenDialog] = useState(false);
  const pendingStatusIdRef = useRef<string | null>(null);

  const status =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  const { userRole } = useUserRole();
  const isGuest = userRole === "guest";
  const isListRow = !asKanban;
  const selectedObjective =
    objectiveId && objective?.id === objectiveId ? objective : null;
  const selectedSprint = sprintId && sprint?.id === sprintId ? sprint : null;
  const isDoneStatus = (statusId: string) => {
    const status = statuses.find((s) => s.id === statusId);
    return status?.category === "completed";
  };

  const getUndoneChildren = () => {
    const activeStatusIds = new Set<string>();
    for (const status of statuses) {
      if (
        status.category === "started" ||
        status.category === "unstarted" ||
        status.category === "backlog"
      ) {
        activeStatusIds.add(status.id);
      }
    }

    const undoneChildren: string[] = [];
    for (const subStory of subStories) {
      if (activeStatusIds.has(subStory.statusId)) {
        undoneChildren.push(subStory.id);
      }
    }

    return undoneChildren;
  };

  const handleStatusUpdate = (statusId: string) => {
    if (isDoneStatus(statusId)) {
      const undoneChildrenIds = getUndoneChildren();
      if (undoneChildrenIds.length > 0) {
        pendingStatusIdRef.current = statusId;
        setShowChildrenDialog(true);
        return; // Don't update yet
      }
    }

    // Normal update if no confirmation needed
    handleUpdate({ statusId });
  };

  const handleConfirmStatusChange = (markChildrenAsDone: boolean) => {
    const pendingStatusId = pendingStatusIdRef.current;
    if (!pendingStatusId) return;

    // Update the main story
    handleUpdate({ statusId: pendingStatusId });

    if (markChildrenAsDone) {
      const undoneChildrenIds = getUndoneChildren();
      for (const childId of undoneChildrenIds) {
        handleUpdate({ id: childId, statusId: pendingStatusId });
      }
    }

    // Reset dialog state
    setShowChildrenDialog(false);
    pendingStatusIdRef.current = null;
  };

  return (
    <>
      {isColumnVisible("Status") && (
        <StatusesMenu>
          <StatusesMenu.Trigger>
            <Button
              className="gap-1 pr-2"
              disabled={isGuest}
              rounded="md"
              size="xs"
              style={{
                backgroundColor: hexToRgba(status?.color, 0.1),
                borderColor: hexToRgba(status?.color, 0.2),
              }}
              type="button"
              variant="outline"
            >
              <StoryStatusIcon statusId={statusId} />
              {status?.name}
            </Button>
          </StatusesMenu.Trigger>
          <StatusesMenu.Items
            setStatusId={handleStatusUpdate}
            statusId={statusId}
            teamId={teamId}
          />
        </StatusesMenu>
      )}
      {isColumnVisible("Priority") && (
        <PrioritiesMenu>
          <PrioritiesMenu.Trigger>
            {asKanban ? (
              <Button
                className="gap-1 pr-2"
                color="tertiary"
                disabled={isGuest}
                size="xs"
                type="button"
                variant="outline"
              >
                <PriorityIcon className="h-[1.15rem]" priority={priority} />
                {priority}
              </Button>
            ) : (
              <button
                aria-label={priority}
                className="flex items-center gap-1 select-none disabled:cursor-not-allowed disabled:opacity-50"
                disabled={isGuest}
                type="button"
              >
                <PriorityIcon priority={priority} />
                <span className={cn("hidden", { "@6xl:inline": isListRow })}>
                  {priority}
                </span>
              </button>
            )}
          </PrioritiesMenu.Trigger>
          <PrioritiesMenu.Items
            priority={priority}
            setPriority={(p) => {
              handleUpdate({ priority: p });
            }}
          />
        </PrioritiesMenu>
      )}
      {isColumnVisible("Estimate") && estimateValue ? (
        <EstimateMenu>
          <Tooltip
            className="pointer-events-none py-3"
            title={formatEstimate(estimateScheme, estimateValue, "full")}
          >
            <span>
              <EstimateMenu.Trigger>
                <Button
                  aria-label={formatEstimate(
                    estimateScheme,
                    estimateValue,
                    "full",
                  )}
                  className="gap-1 px-2"
                  color="tertiary"
                  disabled={isGuest}
                  rounded="md"
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  <EstimateIcon className="h-4" />
                  <span>{formatEstimate(estimateScheme, estimateValue)}</span>
                </Button>
              </EstimateMenu.Trigger>
            </span>
          </Tooltip>
          <EstimateMenu.Items
            estimateScheme={estimateScheme}
            estimateValue={estimateValue}
            setEstimateValue={(estimateValue) => {
              handleUpdate({ estimateValue });
            }}
          />
        </EstimateMenu>
      ) : null}
      {isColumnVisible("Objective") && selectedObjective ? (
        <ObjectivesMenu>
          <Tooltip
            className="w-80 max-w-[calc(100vw-2rem)] py-3"
            collisionPadding={16}
            title={
              <Flex align="start" gap={2}>
                <ObjectiveIcon className="relative top-[3px] h-4 shrink-0" />
                <Box className="min-w-0">
                  <Text className="mb-1.5" fontSize="md">
                    {selectedObjective.name}
                  </Text>
                  <Box
                    className="text-text-muted mt-1 line-clamp-4 min-w-0 break-words"
                    html={selectedObjective.description ?? ""}
                  />
                </Box>
              </Flex>
            }
          >
            <span>
              <ObjectivesMenu.Trigger>
                <Button
                  aria-label={selectedObjective.name}
                  className="gap-1 px-2"
                  color="tertiary"
                  disabled={isGuest}
                  rounded="md"
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  <ObjectiveIcon className="h-4" />
                  <span
                    className={cn("max-w-32 truncate", {
                      "inline-block": asKanban,
                      "hidden @7xl:inline-block": isListRow,
                    })}
                  >
                    {selectedObjective.name}
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
            teamId={teamId}
          />
        </ObjectivesMenu>
      ) : null}
      {isColumnVisible("Sprint") && selectedSprint ? (
        <SprintsMenu>
          <Tooltip
            className="pointer-events-none max-w-96 py-3"
            title={sprintTooltip(selectedSprint)}
          >
            <span>
              <SprintsMenu.Trigger>
                <Button
                  aria-label={selectedSprint.name}
                  className="gap-1 px-2"
                  color="tertiary"
                  disabled={isGuest}
                  rounded="md"
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  <SprintsIcon className="relative -top-[0.3px] h-[1.1rem]" />
                  <span
                    className={cn("max-w-36 truncate", {
                      "inline-block": asKanban,
                      "hidden @7xl:inline-block": isListRow,
                    })}
                  >
                    {selectedSprint.name}
                  </span>
                </Button>
              </SprintsMenu.Trigger>
            </span>
          </Tooltip>
          <SprintsMenu.Items
            setSprintId={(sprintId) => {
              handleUpdate({ sprintId });
            }}
            sprintId={sprintId ?? undefined}
            teamId={teamId}
          />
        </SprintsMenu>
      ) : null}
      {subStories.length > 0 && (
        <Tooltip
          title={
            <Box className="min-w-72">
              <Text
                className="mb-2.5 flex items-center gap-2 pt-1"
                fontSize="md"
              >
                <SubStoryIcon />
                Sub{" "}
                {getTermDisplay("storyTerm", {
                  variant: "plural",
                })}
              </Text>
              <Divider />
              {subStories.map((subStory, idx) => (
                <Link
                  href={withWorkspace(
                    `/story/${subStory.id}/${slugify(subStory.title)}`,
                  )}
                  key={subStory.id}
                >
                  <RowWrapper
                    className={cn(
                      "group border-border max-w-72 gap-4 px-0 py-2 md:px-0",
                      {
                        "border-b-0": idx === subStories.length - 1,
                      },
                    )}
                  >
                    <Flex align="center" gap={2}>
                      <Text color="muted">
                        {teamCode}-{subStory.sequenceId}
                      </Text>
                      <Text className="line-clamp-1 group-hover:underline">
                        {subStory.title}
                      </Text>
                    </Flex>
                    <Flex align="center" className="shrink-0" gap={2}>
                      <PriorityIcon priority={subStory.priority} />
                      <StoryStatusIcon
                        className="h-[1.15rem]"
                        statusId={subStory.statusId}
                      />
                    </Flex>
                  </RowWrapper>
                </Link>
              ))}
            </Box>
          }
        >
          <Badge
            className="text-foreground hidden h-[1.85rem] cursor-pointer bg-transparent text-[0.95rem] md:flex"
            color="tertiary"
            onClick={() => {
              if (!asKanban) {
                setIsExpanded?.(!isExpanded);
              }
            }}
            role="button"
            rounded="md"
            tabIndex={0}
          >
            <SubStoryIcon />
            {subStories.length <= 10 ? subStories.length : `10+`} sub{" "}
            {getTermDisplay("storyTerm", {
              variant: subStories.length === 1 ? "singular" : "plural",
            })}
            {!asKanban && (
              <ArrowRight2Icon
                className={cn("h-4 transition-transform", {
                  "rotate-90": isExpanded,
                })}
                strokeWidth={3}
              />
            )}
          </Badge>
        </Tooltip>
      )}
      {isColumnVisible("Labels") && storyLabels && storyLabels.length > 0 ? (
        <Labels
          compact={isListRow}
          isRectangular={asKanban}
          storyId={id}
          storyLabels={storyLabels}
          teamId={teamId}
        />
      ) : null}
      {isColumnVisible("Deadline") &&
      endDate &&
      !completedOrCancelled(status?.category) ? (
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
                  className={cn("pr-2", {
                    "text-primary dark:text-primary":
                      new Date(endDate) < new Date(),
                    "text-warning dark:text-warning":
                      new Date(endDate) <= addDays(new Date(), 7) &&
                      new Date(endDate) >= new Date(),
                    "px-2": !asKanban,
                  })}
                  color="tertiary"
                  disabled={isGuest}
                  rounded="md"
                  size="xs"
                  type="button"
                  variant="outline"
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
            onDayClick={(day) => {
              handleUpdate({
                endDate: formatISO(day, {
                  representation: "date",
                }),
              });
            }}
            selected={new Date(endDate)}
          />
        </DatePicker>
      ) : null}
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
      <ConfirmDialog
        cancelText="No, leave as is"
        confirmText="Yes, mark as done"
        description={`You're about to mark this ${getTermDisplay(
          "storyTerm",
        )} as done. This ${getTermDisplay(
          "storyTerm",
        )} has sub-${getTermDisplay("storyTerm", {
          variant: subStories.length > 1 ? "plural" : "singular",
        })} that are still in progress. Would you like to mark all sub-${getTermDisplay(
          "storyTerm",
          { variant: subStories.length > 1 ? "plural" : "singular" },
        )} as done as well?`}
        hideClose
        isOpen={showChildrenDialog}
        onCancel={() => {
          handleConfirmStatusChange(false);
        }}
        onConfirm={() => {
          handleConfirmStatusChange(true);
        }}
        title={`Mark sub-${getTermDisplay("storyTerm", {
          variant: subStories.length > 1 ? "plural" : "singular",
        })} as done too?`}
      />
    </>
  );
};
