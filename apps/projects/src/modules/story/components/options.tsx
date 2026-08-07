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
import {
  CalendarIcon,
  EstimateIcon,
  ObjectiveIcon,
  OKRIcon,
  SprintsIcon,
  TagsIcon,
  UsersAddIcon,
} from "icons";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { formatEstimate } from "@/lib/estimate";
import { useStatuses } from "@/lib/hooks/statuses";
import { useStoryById } from "@/modules/story/hooks/story";
import {
  PrioritiesMenu,
  StatusesMenu,
  AssigneesMenu,
  SprintsMenu,
  EstimateMenu,
  StoryStatusIcon,
  PriorityIcon,
  LabelsMenu,
  CollaboratorsMenu,
  StoryLabel,
  ConfirmDialog,
} from "@/components/ui";
import {
  KeyResultMenu,
  ObjectiveKeyResultMenu,
} from "@/components/ui/story/objective-key-result-menu";
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
import { useSprint } from "@/modules/sprints/hooks/sprint-details";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { useKeyResults } from "@/modules/objectives/hooks";
import { useUpdateStoryMutation } from "../hooks/update-mutation";
import type { DetailedStory } from "../types";
import { useUpdateLabelsMutation } from "../hooks/update-labels-mutation";
import { useUpdateCollaboratorsMutation } from "../hooks/collaboration-mutations";
import { OptionsHeader } from "./options-header";

export const Option = ({
  label,
  value,
  className,
  isCompact = false,
  isNotifications,
}: {
  label: string;
  value: ReactNode;
  className?: string;
  isCompact?: boolean;
  isNotifications: boolean;
}) => {
  const isMobile = useMediaQuery("(max-width: 768px)");
  if (isMobile || isCompact) {
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
  variant = "sidebar",
}: {
  storyId: string;
  isNotifications: boolean;
  isDialog?: boolean;
  variant?: "sidebar" | "inline";
}) => {
  const { data } = useStoryById(storyId);
  const {
    priority,
    statusId,
    startDate,
    endDate,
    objectiveId,
    keyResultId,
    assigneeId,
    collaboratorIds,
    collaborators,
    reporterId,
    teamId,
    estimateValue,
    estimateScheme,
    labels: storyLabels,
    sprintId,
    deletedAt,
    subStories,
  } = data!;
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const sprintsEnabled = useSprintsEnabled(teamId);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const isCompact = isMobile || variant === "inline";
  const isInline = variant === "inline";
  const [showChildrenDialog, setShowChildrenDialog] = useState(false);
  const [pendingStatusId, setPendingStatusId] = useState<string | null>(null);
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const { data: sprint } = useSprint(sprintId, teamId);
  const { data: objective } = useObjective(objectiveId, teamId);
  const { data: keyResults = [] } = useKeyResults(
    objectiveId ?? "",
    Boolean(objectiveId && keyResultId),
  );
  const keyResult = keyResults.find(({ id }) => id === keyResultId);
  const objectiveName =
    typeof objective?.name === "string" ? objective.name.trim() || null : null;
  const keyResultName =
    typeof keyResult?.name === "string" ? keyResult.name.trim() || null : null;
  const status =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  const name = status?.name;
  const isDeleted = Boolean(deletedAt);
  const assignee = data?.assignee ?? members.find((m) => m.id === assigneeId);
  const collaboratorLookup = new Map(
    members.map(({ id, username, fullName, avatarUrl }) => [
      id,
      { id, username, fullName, avatarUrl },
    ]),
  );
  for (const collaborator of collaborators) {
    collaboratorLookup.set(collaborator.id, collaborator);
  }
  const selectedCollaborators = collaboratorIds.flatMap((id) => {
    const collaborator = collaboratorLookup.get(id);
    return collaborator ? [collaborator] : [];
  });
  const visibleCollaborators = selectedCollaborators.slice(0, 5);
  const hiddenCollaboratorCount =
    collaboratorIds.length - visibleCollaborators.length;
  const singleCollaborator = selectedCollaborators.at(0);
  let collaboratorButtonIcon: ReactNode = (
    <UsersAddIcon className="h-[1.15rem] w-auto" />
  );
  let collaboratorButtonContent: ReactNode = "Collaborators";

  if (collaboratorIds.length === 1) {
    collaboratorButtonIcon = (
      <Avatar
        name={singleCollaborator?.fullName || singleCollaborator?.username}
        size="xs"
        src={singleCollaborator?.avatarUrl}
      />
    );
    collaboratorButtonContent = (
      <span className="max-w-48 truncate">
        {singleCollaborator?.username ||
          singleCollaborator?.fullName ||
          "Collaborator"}
      </span>
    );
  } else if (collaboratorIds.length > 1) {
    collaboratorButtonIcon = null;
    collaboratorButtonContent = (
      <Flex className="-space-x-1.5">
        {visibleCollaborators.map((collaborator) => (
          <Avatar
            className="ring-surface ring-1"
            key={collaborator.id}
            name={collaborator.fullName || collaborator.username}
            size="xs"
            src={collaborator.avatarUrl}
          />
        ))}
        {hiddenCollaboratorCount > 0 ? (
          <span className="bg-surface-muted ring-surface flex size-5 items-center justify-center rounded-full text-xs ring-1">
            +{hiddenCollaboratorCount}
          </span>
        ) : null}
      </Flex>
    );
  }
  const { data: allLabels = [] } = useLabels();
  const labels = allLabels.filter((label) => storyLabels?.includes(label.id));
  const { mutate } = useUpdateStoryMutation();
  const { mutate: updateLabels } = useUpdateLabelsMutation();
  const { mutate: updateCollaborators } = useUpdateCollaboratorsMutation();
  const { isAdminOrOwner } = useIsAdminOrOwner(reporterId);
  const { userRole } = useUserRole();
  const isGuest = userRole === "guest";

  // References to button elements for keyboard shortcuts
  const statusButtonRef = useRef<HTMLButtonElement>(null);
  const priorityButtonRef = useRef<HTMLButtonElement>(null);
  const assigneeButtonRef = useRef<HTMLButtonElement>(null);
  const estimateButtonRef = useRef<HTMLButtonElement>(null);
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

  useHotkeys("e", (e) => {
    e.preventDefault();
    if (!isDeleted && !isGuest) {
      estimateButtonRef.current?.click();
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
        isInline
          ? "h-auto bg-transparent bg-none p-0 md:h-auto md:overflow-visible md:pb-0"
          : "md:bg-surface-muted/50 dark:md:bg-surface-elevated/40 bg-transparent pb-2 md:h-dvh md:overflow-y-auto md:pb-6",
        {
          "dark:md:bg-surface-elevated/80 h-[85dvh]": isDialog,
        },
      )}
    >
      {!isInline ? (
        <Box className="hidden md:block">
          <OptionsHeader
            isAdminOrOwner={isAdminOrOwner}
            isDialog={isDialog}
            storyId={storyId}
          />
        </Box>
      ) : null}
      <Container
        className={cn("text-text-muted px-0.5 pt-4 md:px-6", {
          "px-0 pt-0 md:px-0": isInline,
        })}
      >
        <Box
          className={cn(
            "mb-0 grid grid-cols-[9rem_auto] items-center gap-3 md:mb-6",
            {
              "hidden md:mb-0": isInline && !isDeleted,
            },
          )}
        >
          {!isNotifications && (
            <Text className="hidden md:block" fontWeight="semibold">
              Properties
            </Text>
          )}
          {isDeleted ? (
            <Badge
              className="text-foreground border-opacity-30 bg-opacity-30 px-2"
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
        <Box
          className={cn("flex flex-wrap gap-2", {
            "md:block": !isCompact,
          })}
        >
          <Option
            isCompact={isCompact}
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
                    size="sm"
                    type="button"
                    variant={isCompact ? "solid" : "naked"}
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
            isCompact={isCompact}
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
                    size="sm"
                    type="button"
                    variant={isCompact ? "solid" : "naked"}
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
            isCompact={isCompact}
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
                          "text-foreground/80": !assignee?.fullName,
                        })}
                        name={assignee?.fullName}
                        size="xs"
                        src={assignee?.avatarUrl}
                      />
                    }
                    ref={assigneeButtonRef}
                    size="sm"
                    type="button"
                    variant={isCompact ? "solid" : "naked"}
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
            isCompact={isCompact}
            isNotifications={isNotifications}
            label="Collaborators"
            value={
              <CollaboratorsMenu>
                <CollaboratorsMenu.Trigger>
                  <Button
                    className={cn("max-w-full font-medium", {
                      "text-text-muted": collaboratorIds.length === 0,
                    })}
                    color="tertiary"
                    disabled={isDeleted || isGuest}
                    leftIcon={collaboratorButtonIcon}
                    size="sm"
                    type="button"
                    variant={isCompact ? "solid" : "naked"}
                  >
                    {collaboratorButtonContent}
                  </Button>
                </CollaboratorsMenu.Trigger>
                <CollaboratorsMenu.Items
                  assigneeId={assigneeId}
                  collaboratorIds={collaboratorIds}
                  onCollaboratorsChange={(collaboratorIds) => {
                    updateCollaborators({ storyId, collaboratorIds });
                  }}
                  teamId={teamId}
                />
              </CollaboratorsMenu>
            }
          />
          <Option
            isCompact={isCompact}
            isNotifications={isNotifications}
            label="Estimate"
            value={
              <EstimateMenu>
                <EstimateMenu.Trigger>
                  <Button
                    className={cn("font-medium", {
                      "text-text-muted": !estimateValue,
                    })}
                    color="tertiary"
                    disabled={isDeleted || isGuest}
                    leftIcon={
                      <EstimateIcon
                        className={cn("h-[1.15rem] w-auto", {
                          "text-text-muted": !estimateValue,
                        })}
                      />
                    }
                    ref={estimateButtonRef}
                    size="sm"
                    type="button"
                    variant={isCompact ? "solid" : "naked"}
                  >
                    {estimateValue ? (
                      formatEstimate(estimateScheme, estimateValue, "full")
                    ) : (
                      <Text as="span" color="muted">
                        Add estimate
                      </Text>
                    )}
                  </Button>
                </EstimateMenu.Trigger>
                <EstimateMenu.Items
                  estimateScheme={estimateScheme}
                  estimateValue={estimateValue}
                  setEstimateValue={(estimateValue) => {
                    handleUpdate({ estimateValue });
                  }}
                />
              </EstimateMenu>
            }
          />
          <Option
            isCompact={isCompact}
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
                    variant={isCompact ? "solid" : "naked"}
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
            isCompact={isCompact}
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
                      <Box>
                        {getDueDateMessage(
                          new Date(endDate!),
                          getTermDisplay("storyTerm"),
                        )}
                      </Box>
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
                        variant={isCompact ? "solid" : "naked"}
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
            <>
              <Option
                isCompact={isCompact}
                isNotifications={isNotifications}
                label={getTermDisplay("objectiveTerm", { capitalize: true })}
                value={
                  <ObjectiveKeyResultMenu
                    align="end"
                    keyResultId={keyResultId}
                    objectiveId={objectiveId}
                    onChange={(selection) => {
                      handleUpdate(selection);
                    }}
                    teamId={teamId}
                  >
                    <Button
                      className="w-fit max-w-[13rem] justify-start font-medium"
                      color="tertiary"
                      disabled={isDeleted || isGuest}
                      leftIcon={
                        <ObjectiveIcon
                          className={cn("h-[1.15rem] w-auto shrink-0", {
                            "text-text-muted": !objectiveId,
                          })}
                        />
                      }
                      ref={objectiveButtonRef}
                      size="sm"
                      title={objectiveName ?? undefined}
                      type="button"
                      variant={isCompact ? "solid" : "naked"}
                    >
                      <span className="block min-w-0 truncate">
                        {objectiveId
                          ? objectiveName ??
                            getTermDisplay("objectiveTerm", {
                              capitalize: true,
                            })
                          : `Add ${getTermDisplay("objectiveTerm")}`}
                      </span>
                    </Button>
                  </ObjectiveKeyResultMenu>
                }
              />
              {objectiveId ? (
                <Option
                  isCompact={isCompact}
                  isNotifications={isNotifications}
                  label={getTermDisplay("keyResultTerm", {
                    capitalize: true,
                  })}
                  value={
                    <KeyResultMenu
                      align="end"
                      keyResultId={keyResultId}
                      objectiveId={objectiveId}
                      onChange={(nextKeyResultId) => {
                        handleUpdate({ keyResultId: nextKeyResultId });
                      }}
                    >
                      <Button
                        className="w-fit max-w-[13rem] justify-start font-medium"
                        color="tertiary"
                        disabled={isDeleted || isGuest}
                        leftIcon={
                          <OKRIcon
                            className={cn("h-[1.15rem] w-auto shrink-0", {
                              "text-text-muted": !keyResultId,
                            })}
                            strokeWidth={2.4}
                          />
                        }
                        size="sm"
                        title={keyResultName ?? undefined}
                        type="button"
                        variant={isCompact ? "solid" : "naked"}
                      >
                        <span className="block min-w-0 truncate">
                          {keyResultId
                            ? keyResultName ??
                              getTermDisplay("keyResultTerm", {
                                capitalize: true,
                              })
                            : `Add ${getTermDisplay("keyResultTerm")}`}
                        </span>
                      </Button>
                    </KeyResultMenu>
                  }
                />
              ) : null}
            </>
          ) : null}
          {sprintsEnabled ? (
            <Option
              isCompact={isCompact}
              isNotifications={isNotifications}
              label="Sprint"
              value={
                <SprintsMenu>
                  <SprintsMenu.Trigger>
                    <Button
                      color="tertiary"
                      disabled={isDeleted || isGuest}
                      leftIcon={
                        <SprintsIcon
                          className={cn("h-5 w-auto", {
                            "text-text-muted": !sprintId,
                          })}
                        />
                      }
                      ref={sprintButtonRef}
                      size="sm"
                      type="button"
                      variant={isCompact ? "solid" : "naked"}
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
            isCompact={isCompact}
            isNotifications={isNotifications}
            label="Labels"
            value={
              <Box
                className={cn({
                  "md:ml-2.5": !isCompact,
                  "md:ml-0": !isCompact && labels.length === 0,
                })}
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
                            leftIcon={<TagsIcon className="h-4 w-auto" />}
                            ref={labelsButtonRef}
                            rounded="full"
                            size="sm"
                            title="Add labels"
                            type="button"
                            variant={isCompact ? "solid" : "naked"}
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
                        leftIcon={<TagsIcon className="h-[1.15rem] w-auto" />}
                        ref={emptyLabelsButtonRef}
                        size="sm"
                        type="button"
                        variant={isCompact ? "solid" : "naked"}
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

        {!isInline ? <Divider className="my-4" /> : null}
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
