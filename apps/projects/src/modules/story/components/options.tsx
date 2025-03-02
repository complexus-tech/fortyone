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
import { type ReactNode } from "react";
import { addDays, format, differenceInDays, formatISO } from "date-fns";
import { CalendarIcon, ObjectiveIcon, PlusIcon, SprintsIcon } from "icons";
import { cn } from "lib";
import { useParams } from "next/navigation";
import Link from "next/link";
import { useStatuses } from "@/lib/hooks/statuses";
import { useMembers } from "@/lib/hooks/members";
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
import { useSprints } from "@/modules/sprints/hooks/sprints";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useUserRole } from "@/hooks";
import { useUpdateStoryMutation } from "../hooks/update-mutation";
import type { DetailedStory } from "../types";
import { useUpdateLabelsMutation } from "../hooks/update-labels-mutation";
import { AddLinks, OptionsHeader } from ".";

export const Option = ({
  label,
  value,
  className,
}: {
  label: string;
  value: ReactNode;
  className?: string;
}) => {
  return (
    <Box
      className={cn(
        "my-4 grid grid-cols-[9rem_auto] items-center gap-3",
        className,
      )}
    >
      <Text
        className="flex items-center gap-1 truncate"
        color="muted"
        fontWeight="medium"
      >
        {label}
      </Text>
      {value}
    </Box>
  );
};

export const Options = () => {
  const params = useParams<{ storyId: string }>();
  const { data } = useStoryById(params.storyId);
  const {
    id,
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
  const labels = allLabels.filter((label) => storyLabels.includes(label.id));
  const { mutate } = useUpdateStoryMutation();
  const { mutate: updateLabels } = useUpdateLabelsMutation();
  const { isAdminOrOwner } = useIsAdminOrOwner(reporterId);
  const { userRole } = useUserRole();
  const isGuest = userRole === "guest";

  const handleUpdate = (data: Partial<DetailedStory>) => {
    mutate({ storyId: id, payload: data });
  };

  const handleUpdateLabels = (labels: string[] = []) => {
    updateLabels({ storyId: id, labels });
  };

  return (
    <Box className="h-full overflow-y-auto bg-gradient-to-br from-white via-gray-50/50 to-gray-50 pb-6 dark:from-dark-200/50 dark:to-dark">
      <OptionsHeader isAdminOrOwner={isAdminOrOwner} />
      <Container className="pt-4 text-gray-300/90 md:px-6">
        <Box className="mb-6 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text fontWeight="semibold">Properties</Text>
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
        <Option
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
        <Option
          label="Status"
          value={
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={isDeleted || isGuest}
                  leftIcon={<StoryStatusIcon statusId={statusId} />}
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
          label="Priority"
          value={
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={isDeleted || isGuest}
                  leftIcon={<PriorityIcon priority={priority} />}
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
              />
            </AssigneesMenu>
          }
        />
        <Option
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
          label="Due date"
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
                      variant="naked"
                    >
                      {endDate ? (
                        format(new Date(endDate), "MMM d, yyyy")
                      ) : (
                        <Text color="muted">Add due date</Text>
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
        <Option
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
              />
            </ObjectivesMenu>
          }
        />
        <Option
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
              />
            </SprintsMenu>
          }
        />
        <Option
          className={cn("items-start pt-1", {
            "items-center pt-0": labels.length === 0,
          })}
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
                        labelIds={storyLabels}
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
                        labelIds={storyLabels}
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
                        labelIds={storyLabels}
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
                      type="button"
                      variant="naked"
                    >
                      Add labels
                    </Button>
                  </LabelsMenu.Trigger>
                  <LabelsMenu.Items
                    labelIds={storyLabels}
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
        <AddLinks storyId={id} />
      </Container>
    </Box>
  );
};
