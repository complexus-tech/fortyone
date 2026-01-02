"use client";
import {
  Box,
  Button,
  Container,
  Divider,
  Text,
  DatePicker,
  Avatar,
  Flex,
  Tooltip,
  Badge,
} from "ui";
import { type ReactNode, useRef, useState } from "react";
import { addDays, format, differenceInDays, formatISO } from "date-fns";
import { CalendarIcon, ObjectiveIcon, PlusIcon, SprintsIcon } from "icons";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { useStatuses } from "@/lib/hooks/statuses";
import { useStoryById } from "@/modules/story/hooks/story";
import {
  PrioritiesMenu,
  StatusesMenu,
  AssigneesMenu,
  SprintsMenu,
  StoryStatusIcon,
  PriorityIcon,
  LabelsMenu,
  StoryLabel,
  ConfirmDialog,
} from "@/components/ui";
import { ObjectivesMenu } from "@/components/ui/story/objectives-menu";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useLabels } from "@/lib/hooks/labels";
import { getDueDateMessage } from "@/components/ui/story/due-date-tooltip";
import { useIsAdminOrOwner } from "@/hooks/owner";
import {
  useFeatures,
  useMediaQuery,
  useTerminology,
  useUserRole,
  useSprintsEnabled,
} from "@/hooks";
import { useMembers } from "@/lib/hooks/members";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import { useSprint } from "@/modules/sprints/hooks/sprint-details";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { useUpdateStoryMutation } from "../hooks/update-mutation";
import type { DetailedStory } from "../types";
import { useUpdateLabelsMutation } from "../hooks/update-labels-mutation";
import { AddLinks, OptionsHeader } from ".";

export const Option = ({
  label,
  value,
  className,
  isNotifications,
}: {
  label: string;
  value: ReactNode;
  className?: string;
  isNotifications: boolean;
}) => {
  const isMobile = useMediaQuery("(max-width: 768px)");
  if (isMobile) {
    return value;
  }
  return (
    <Box
      className={cn(
        "my-4 grid grid-cols-[7.5rem_auto] items-center gap-3 md:my-5",
        { "grid-cols-1": isNotifications },
        className,
      )}
    >
      {!isNotifications && (
        <Text
          className="flex items-center gap-1 truncate"
          color="muted"
          fontWeight="medium"
        >
          {label}
        </Text>
      )}
      {value}
    </Box>
  );
};

export const Options = ({
  storyId,
  isNotifications,
  isDialog,
}: {
  storyId: string;
  isNotifications: boolean;
  isDialog?: boolean;
}) => {
  const { data } = useStoryById(storyId);
  const {
    priority,
    statusId,
    startDate,
    endDate,
    objectiveId,
    assigneeId,
    reporterId,
    teamId,
    labels: storyLabels,
    sprintId,
    deletedAt,
    subStories,
  } = data!;
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const sprintsEnabled = useSprintsEnabled(teamId);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [showChildrenDialog, setShowChildrenDialog] = useState(false);
  const [pendingStatusId, setPendingStatusId] = useState<string | null>(null);
  const { data: sprints = [] } = useTeamSprints(teamId);
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const { data: objectives = [] } = useTeamObjectives(teamId);
  const { data: sprint } = useSprint(sprintId, teamId);
  const { data: objective } = useObjective(objectiveId, teamId);
  const status =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  const name = status?.name;
  const isDeleted = Boolean(deletedAt);
  const assignee = members.find((m) => m.id === assigneeId);
  const { data: allLabels = [] } = useLabels();
  const labels = allLabels.filter((label) => storyLabels?.includes(label.id));
  const { mutate } = useUpdateStoryMutation();
  const { mutate: updateLabels } = useUpdateLabelsMutation();
  const { isAdminOrOwner } = useIsAdminOrOwner(reporterId);
  const { userRole } = useUserRole();
  const isGuest = userRole === "guest";

  // References to button elements for keyboard shortcuts
  const statusButtonRef = useRef<HTMLButtonElement>(null);
  const priorityButtonRef = useRef<HTMLButtonElement>(null);
  const assigneeButtonRef = useRef<HTMLButtonElement>(null);
  const startDateButtonRef = useRef<HTMLButtonElement>(null);
  const dueDateButtonRef = useRef<HTMLButtonElement>(null);
  const labelsButtonRef = useRef<HTMLButtonElement>(null);
  const emptyLabelsButtonRef = useRef<HTMLButtonElement>(null);
  const objectiveButtonRef = useRef<HTMLButtonElement>(null);
  const sprintButtonRef = useRef<HTMLButtonElement>(null);

  const getUndoneChildren = () => {
    const unstartedAndStartedStatusIds = statuses
      .filter(
        (status) =>
          status.category === "started" ||
          status.category === "unstarted" ||
          status.category === "backlog",
      )
      .map((s) => s.id);

    return subStories
      .filter((subStory) =>
        unstartedAndStartedStatusIds.includes(subStory.statusId),
      )
      .map((s) => s.id);
  };

  const isDoneStatus = (statusId: string) => {
    const status = statuses.find((s) => s.id === statusId);
    return status?.category === "completed";
  };

  const handleUpdate = (data: Partial<DetailedStory>) => {
    // If updating status to a "done" state and has undone children
    if (data.statusId && isDoneStatus(data.statusId)) {
      const undoneChildrenIds = getUndoneChildren();
      if (undoneChildrenIds.length > 0) {
        setPendingStatusId(data.statusId);
        setShowChildrenDialog(true);
        return; // Don't update yet, wait for user confirmation
      }
    }

    // Normal update if no confirmation needed
    mutate({ storyId, payload: data });
  };

  const handleUpdateLabels = (labels: string[] = []) => {
    updateLabels({ storyId, labels });
  };

  const handleConfirmStatusChange = (markChildrenAsDone: boolean) => {
    if (!pendingStatusId) return;

    // Update the main story
    mutate({ storyId, payload: { statusId: pendingStatusId } });

    if (markChildrenAsDone) {
      const undoneChildrenIds = getUndoneChildren();
      // Update all undone children to the same status
      for (const childId of undoneChildrenIds) {
        mutate({ storyId: childId, payload: { statusId: pendingStatusId } });
      }
    }
    // Reset dialog state
    setShowChildrenDialog(false);
    setPendingStatusId(null);
  };

  useHotkeys("s", (e) => {
    e.preventDefault();
    if (!isDeleted && !isGuest) {
      statusButtonRef.current?.click();
    }
  });

  useHotkeys("p", (e) => {
    e.preventDefault();
    if (!isDeleted && !isGuest) {
      priorityButtonRef.current?.click();
    }
  });

  useHotkeys("a", (e) => {
    e.preventDefault();
    if (!isDeleted && !isGuest) {
      assigneeButtonRef.current?.click();
    }
  });

  useHotkeys("d", (e) => {
    e.preventDefault();
    if (!isDeleted && !isGuest) {
      dueDateButtonRef.current?.click();
    }
  });

  useHotkeys("l", (e) => {
    e.preventDefault();
    if (!isDeleted && !isGuest) {
      if (labels.length > 0) {
        labelsButtonRef.current?.click();
      } else {
        emptyLabelsButtonRef.current?.click();
      }
    }
  });

  useHotkeys("o", (e) => {
    e.preventDefault();
    if (!isDeleted && !isGuest && features.objectiveEnabled) {
      objectiveButtonRef.current?.click();
    }
  });

  useHotkeys("n", (e) => {
    e.preventDefault();
    if (!isDeleted && !isGuest && sprintsEnabled) {
      sprintButtonRef.current?.click();
    }
  });

  useHotkeys("b", (e) => {
    e.preventDefault();
    if (!isDeleted && !isGuest) {
      startDateButtonRef.current?.click();
    }
  });

  return (
    <Box
      className={cn(
        "bg-surface to-surface-muted pb-2 md:h-dvh md:overflow-y-auto md:pb-6",
        {
          "h-[85dvh]": isDialog,
        },
      )}
    >
      <Box className="hidden md:block">
        <OptionsHeader
          isAdminOrOwner={isAdminOrOwner}
          isDialog={isDialog}
          isNotifications={isNotifications}
          storyId={storyId}
        />
      </Box>
      <Container className="px-0.5 pt-4 text-text-muted md:px-6">
        <Box className="mb-0 grid grid-cols-[9rem_auto] items-center gap-3 md:mb-6">
          {!isNotifications && (
            <Text className="hidden md:block" fontWeight="semibold">
              Properties
            </Text>
          )}
          {isDeleted ? (
            <Badge
              className="text-foreground border-opacity-30 px-2 bg-opacity-30"
              color="tertiary"
              size="lg"
            >
              {differenceInDays(
                addDays(new Date(deletedAt!), 30),
                new Date(deletedAt!),
              )}{" "}
              days left in bin
            </Badge>
          ) : null}
        </Box>
        <Box className="flex flex-wrap gap-2 md:block">
          <Option
            isNotifications={isNotifications}
            label="Status"
            value={
              <StatusesMenu>
                <StatusesMenu.Trigger>
                  <Button
                    color="tertiary"
                    disabled={isDeleted || isGuest}
                    leftIcon={<StoryStatusIcon statusId={statusId} />}
                    ref={statusButtonRef}
                    type="button"
                    size="sm"
                    variant={isMobile ? "solid" : "naked"}
                  >
                    {name}
                  </Button>
                </StatusesMenu.Trigger>
                <StatusesMenu.Items
                  setStatusId={(statusId) => {
                    handleUpdate({ statusId });
                  }}
                  statusId={statusId}
                  teamId={teamId}
                />
              </StatusesMenu>
            }
          />
          <Option
            isNotifications={isNotifications}
            label="Priority"
            value={
              <PrioritiesMenu>
                <PrioritiesMenu.Trigger>
                  <Button
                    color="tertiary"
                    disabled={isDeleted || isGuest}
                    leftIcon={<PriorityIcon priority={priority} />}
                    ref={priorityButtonRef}
                    type="button"
                    size="sm"
                    variant={isMobile ? "solid" : "naked"}
                  >
                    {priority}
                  </Button>
                </PrioritiesMenu.Trigger>
                <PrioritiesMenu.Items
                  setPriority={(priority) => {
                    handleUpdate({ priority });
                  }}
                />
              </PrioritiesMenu>
            }
          />
          <Option
            isNotifications={isNotifications}
            label="Assignee"
            value={
              <AssigneesMenu>
                <AssigneesMenu.Trigger>
                  <Button
                    className={cn("font-medium", {
                      "text-text-muted": !assigneeId,
                    })}
                    color="tertiary"
                    disabled={isDeleted || isGuest}
                    leftIcon={
                      <Avatar
                        className={cn({
                          "text-foreground/80":
                            !assignee?.fullName,
                        })}
                        name={assignee?.fullName}
                        size="xs"
                        src={assignee?.avatarUrl}
                      />
                    }
                    ref={assigneeButtonRef}
                    type="button"
                    size="sm"
                    variant={isMobile ? "solid" : "naked"}
                  >
                    {assignee?.username || (
                      <Text as="span" color="muted">
                        Assign
                      </Text>
                    )}
                  </Button>
                </AssigneesMenu.Trigger>
                <AssigneesMenu.Items
                  assigneeId={assigneeId}
                  onAssigneeSelected={(assigneeId) => {
                    handleUpdate({ assigneeId });
                  }}
                  teamId={teamId}
                />
              </AssigneesMenu>
            }
          />
          <Option
            isNotifications={isNotifications}
            label="Start date"
            value={
              <DatePicker>
                <DatePicker.Trigger>
                  <Button
                    color="tertiary"
                    disabled={isDeleted || isGuest}
                    leftIcon={
                      <CalendarIcon
                        className={cn("h-[1.15rem] w-auto", {
                          "text-text-muted": !startDate,
                        })}
                      />
                    }
                    ref={startDateButtonRef}
                    size="sm"
                    variant={isMobile ? "solid" : "naked"}
                  >
                    {startDate ? (
                      format(new Date(startDate), "MMM d, yyyy")
                    ) : (
                      <Text color="muted">Add start date</Text>
                    )}
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  onDayClick={(day) => {
                    handleUpdate({
                      startDate: formatISO(day, { representation: "date" }),
                    });
                  }}
                  selected={startDate ? new Date(startDate) : undefined}
                />
              </DatePicker>
            }
          />
          <Option
            isNotifications={isNotifications}
            label="Deadline"
            value={
              <DatePicker>
                <Tooltip
                  className="py-3"
                  hidden={!endDate}
                  title={
                    <Flex align="start" gap={2}>
                      <CalendarIcon
                        className={cn("relative top-[2.5px] h-5 w-auto", {
                          "text-primary dark:text-primary":
                            new Date(endDate!) < new Date(),
                          "text-warning dark:text-warning":
                            new Date(endDate!) <= addDays(new Date(), 7) &&
                            new Date(endDate!) >= new Date(),
                        })}
                      />
                      <Box>{getDueDateMessage(new Date(endDate!))}</Box>
                    </Flex>
                  }
                >
                  <span>
                    <DatePicker.Trigger>
                      <Button
                        className={cn({
                          "text-primary dark:text-primary":
                            endDate && new Date(endDate) < new Date(),
                          "text-warning dark:text-warning":
                            endDate &&
                            new Date(endDate) <= addDays(new Date(), 7) &&
                            new Date(endDate) >= new Date(),
                          "text-text-muted": !endDate,
                        })}
                        color="tertiary"
                        disabled={isDeleted || isGuest}
                        leftIcon={
                          <CalendarIcon className="h-[1.15rem] w-auto" />
                        }
                        ref={dueDateButtonRef}
                        size="sm"
                        variant={isMobile ? "solid" : "naked"}
                      >
                        {endDate ? (
                          format(new Date(endDate), "MMM d, yyyy")
                        ) : (
                          <Text color="muted">Add deadline</Text>
                        )}
                      </Button>
                    </DatePicker.Trigger>
                  </span>
                </Tooltip>
                <DatePicker.Calendar
                  onDayClick={(day) => {
                    handleUpdate({
                      endDate: formatISO(day, { representation: "date" }),
                    });
                  }}
                  selected={endDate ? new Date(endDate) : undefined}
                />
              </DatePicker>
            }
          />
          {features.objectiveEnabled ? (
            <Option
              isNotifications={isNotifications}
              label="Objective"
              value={
                <ObjectivesMenu>
                  <ObjectivesMenu.Trigger>
                    <Button
                      color="tertiary"
                      disabled={isDeleted || isGuest}
                      leftIcon={
                        objectiveId ? (
                          <ObjectiveIcon className="h-[1.15rem] w-auto" />
                        ) : (
                          <PlusIcon className="h-5 w-auto" />
                        )
                      }
                      ref={objectiveButtonRef}
                      title={objectiveId ? objective?.name : undefined}
                      type="button"
                      size="sm"
                      variant={isMobile ? "solid" : "naked"}
                    >
                      <span className="inline-block max-w-[12ch] truncate">
                        {objective?.name || "Add objective"}
                      </span>
                    </Button>
                  </ObjectivesMenu.Trigger>
                  <ObjectivesMenu.Items
                    align="end"
                    objectiveId={objectiveId ?? undefined}
                    setObjectiveId={(objectiveId) => {
                      handleUpdate({ objectiveId });
                    }}
                    teamId={teamId}
                  />
                </ObjectivesMenu>
              }
            />
          ) : null}
          {sprintsEnabled ? (
            <Option
              isNotifications={isNotifications}
              label="Sprint"
              value={
                <SprintsMenu>
                  <SprintsMenu.Trigger>
                    <Button
                      color="tertiary"
                      disabled={isDeleted || isGuest}
                      leftIcon={
                        sprintId ? (
                          <SprintsIcon className="h-5 w-auto" />
                        ) : (
                          <PlusIcon className="h-5 w-auto" />
                        )
                      }
                      ref={sprintButtonRef}
                      type="button"
                      size="sm"
                      variant={isMobile ? "solid" : "naked"}
                    >
                      <span className="inline-block max-w-[16ch] truncate">
                        {sprint?.name || "Add sprint"}
                      </span>
                    </Button>
                  </SprintsMenu.Trigger>
                  <SprintsMenu.Items
                    align="end"
                    setSprintId={(sprintId) => {
                      handleUpdate({ sprintId });
                    }}
                    sprintId={sprintId ?? undefined}
                    teamId={teamId}
                  />
                </SprintsMenu>
              }
            />
          ) : null}
          <Option
            className={cn("items-start pt-1", {
              "items-center pt-0": labels.length === 0,
            })}
            isNotifications={isNotifications}
            label="Labels"
            value={
              <Box
                className={cn("md:ml-2.5", { "md:ml-0": labels.length === 0 })}
              >
                {labels.length > 0 ? (
                  <Flex align="center" className="gap-1.5" wrap>
                    {labels.slice(0, labels.length - 1).map((label) => (
                      <LabelsMenu key={label.id}>
                        <LabelsMenu.Trigger>
                          <span
                            className={cn({
                              "pointer-events-none cursor-not-allowed":
                                isDeleted || isGuest,
                            })}
                          >
                            <StoryLabel {...label} isRectangular size="sm" />
                          </span>
                        </LabelsMenu.Trigger>
                        <LabelsMenu.Items
                          labelIds={storyLabels ?? []}
                          setLabelIds={(labelIds) => {
                            handleUpdateLabels(labelIds);
                          }}
                          teamId={teamId}
                        />
                      </LabelsMenu>
                    ))}
                    <Flex align="center" gap={1}>
                      <LabelsMenu>
                        <LabelsMenu.Trigger>
                          <span
                            className={cn({
                              "pointer-events-none cursor-not-allowed":
                                isDeleted || isGuest,
                            })}
                          >
                            <StoryLabel
                              {...labels.at(-1)!}
                              isRectangular
                              size="sm"
                            />
                          </span>
                        </LabelsMenu.Trigger>
                        <LabelsMenu.Items
                          labelIds={storyLabels ?? []}
                          setLabelIds={(labelIds) => {
                            handleUpdateLabels(labelIds);
                          }}
                          teamId={teamId}
                        />
                      </LabelsMenu>
                      <LabelsMenu>
                        <LabelsMenu.Trigger>
                          <Button
                            asIcon
                            className="m-0"
                            color="tertiary"
                            disabled={isDeleted || isGuest}
                            leftIcon={<PlusIcon />}
                            ref={labelsButtonRef}
                            rounded="full"
                            size="sm"
                            title="Add labels"
                            type="button"
                            variant={isMobile ? "solid" : "naked"}
                          >
                            <span className="sr-only">Add labels</span>
                          </Button>
                        </LabelsMenu.Trigger>
                        <LabelsMenu.Items
                          labelIds={storyLabels ?? []}
                          setLabelIds={(labelIds) => {
                            handleUpdateLabels(labelIds);
                          }}
                          teamId={teamId}
                        />
                      </LabelsMenu>
                    </Flex>
                  </Flex>
                ) : (
                  <LabelsMenu>
                    <LabelsMenu.Trigger>
                      <Button
                        color="tertiary"
                        disabled={isDeleted || isGuest}
                        leftIcon={<PlusIcon />}
                        ref={emptyLabelsButtonRef}
                        size="sm"
                        type="button"
                        variant={isMobile ? "solid" : "naked"}
                      >
                        Add labels
                      </Button>
                    </LabelsMenu.Trigger>
                    <LabelsMenu.Items
                      labelIds={storyLabels ?? []}
                      setLabelIds={(labelIds) => {
                        handleUpdateLabels(labelIds);
                      }}
                      teamId={teamId}
                    />
                  </LabelsMenu>
                )}
              </Box>
            }
          />
        </Box>

        {/* 
        <Option label="Module" value={<ModulesMenu />} />
        <Option
          label="Parent"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />
        <Option
          label="Blocking"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />
        <Option
          label="Blocked by"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />
        <Option
          label="Related to"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        /> */}

        <Divider className="my-4" />
        <AddLinks storyId={storyId} />
      </Container>

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
    </Box>
  );
};
