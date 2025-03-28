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
import { type ReactNode, useRef } from "react";
import { addDays, format, differenceInDays, formatISO } from "date-fns";
import { CalendarIcon, ObjectiveIcon, PlusIcon, SprintsIcon } from "icons";
import { cn } from "lib";
import Link from "next/link";
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
} from "@/components/ui";
import { ObjectivesMenu } from "@/components/ui/story/objectives-menu";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useLabels } from "@/lib/hooks/labels";
import { getDueDateMessage } from "@/components/ui/story/due-date-tooltip";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useFeatures, useUserRole } from "@/hooks";
import { useSprints } from "@/modules/sprints/hooks/sprints";
import { useMembers } from "@/lib/hooks/members";
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
  return (
    <Box
      className={cn(
        "my-4 grid grid-cols-[9rem_auto] items-center gap-3",
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
}: {
  storyId: string;
  isNotifications: boolean;
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
  } = data!;
  const features = useFeatures();
  const { data: sprints = [] } = useSprints();
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const { data: objectives = [] } = useObjectives();
  const sprint = sprints.find((s) => s.id === sprintId);
  const objective = objectives.find((o) => o.id === objectiveId);
  const { name } = (statuses.find((state) => state.id === statusId) ||
    statuses.at(0))!;
  const isDeleted = Boolean(deletedAt);
  const assignee = members.find((m) => m.id === assigneeId);
  const reporter = members.find((m) => m.id === reporterId);
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

  const handleUpdate = (data: Partial<DetailedStory>) => {
    mutate({ storyId, payload: data });
  };

  const handleUpdateLabels = (labels: string[] = []) => {
    updateLabels({ storyId, labels });
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
    if (!isDeleted && !isGuest && features.sprintEnabled) {
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
    <Box className="h-full overflow-y-auto bg-gradient-to-br from-white via-gray-50/50 to-gray-50 pb-6 dark:from-dark-200/50 dark:to-dark">
      <OptionsHeader
        isAdminOrOwner={isAdminOrOwner}
        isNotifications={isNotifications}
        storyId={storyId}
      />
      <Container className="pt-4 text-gray-300/90 md:px-6">
        <Box className="mb-6 grid grid-cols-[9rem_auto] items-center gap-3">
          {!isNotifications && <Text fontWeight="semibold">Properties</Text>}
          {isDeleted ? (
            <Badge
              className="border-opacity-30 px-2 text-dark dark:bg-opacity-30 dark:text-white"
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
        {!isNotifications && (
          <Option
            isNotifications={isNotifications}
            label="Reporter"
            value={
              <Flex align="center" className="px-2.5" gap={2}>
                <Avatar
                  className={cn({
                    "text-dark/80 dark:text-gray-200": !reporter?.fullName,
                  })}
                  name={reporter?.fullName}
                  size="xs"
                  src={reporter?.avatarUrl}
                />
                <Link
                  className="relative -top-[1px] font-medium text-dark dark:text-gray-200"
                  href={`/profile/${reporter?.id}`}
                >
                  {reporter?.username}
                </Link>
              </Flex>
            }
          />
        )}
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
                  variant="naked"
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
                  variant="naked"
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
                    "text-gray-200 dark:text-gray-300": !assigneeId,
                  })}
                  color="tertiary"
                  disabled={isDeleted || isGuest}
                  leftIcon={
                    <Avatar
                      className={cn({
                        "text-dark/80 dark:text-gray-200": !assignee?.fullName,
                      })}
                      name={assignee?.fullName}
                      size="xs"
                      src={assignee?.avatarUrl}
                    />
                  }
                  ref={assigneeButtonRef}
                  type="button"
                  variant="naked"
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
                        "text-gray/80 dark:text-gray-300/80": !startDate,
                      })}
                    />
                  }
                  ref={startDateButtonRef}
                  variant="naked"
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
                    startDate: formatISO(day),
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
                        "text-gray/80 dark:text-gray-300/80": !endDate,
                      })}
                      color="tertiary"
                      disabled={isDeleted || isGuest}
                      leftIcon={<CalendarIcon className="h-[1.15rem] w-auto" />}
                      ref={dueDateButtonRef}
                      variant="naked"
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
                    endDate: formatISO(day),
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
                    variant="naked"
                  >
                    <span className="inline-block max-w-[16ch] truncate">
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
        {features.sprintEnabled ? (
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
                    variant="naked"
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
            <>
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
                          <StoryLabel {...label} />
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
                          <StoryLabel {...labels.at(-1)!} />
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
                          variant="naked"
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
                      type="button"
                      variant="naked"
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
            </>
          }
        />
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
    </Box>
  );
};
