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
import {
  addDays,
  format,
  differenceInDays,
  isTomorrow,
  formatISO,
} from "date-fns";
import {
  CalendarIcon,
  CloseIcon,
  ObjectiveIcon,
  PlusIcon,
  SprintsIcon,
} from "icons";
import {
  PrioritiesMenu,
  StatusesMenu,
  AssigneesMenu,
  ModulesMenu,
  SprintsMenu,
  StoryStatusIcon,
  PriorityIcon,
} from "@/components/ui";
import { Labels } from "@/components/ui/story/labels";
import { AddLinks, OptionsHeader } from ".";
import { DetailedStory } from "../types";
import { cn } from "lib";
import { useParams } from "next/navigation";
import { useStoryById } from "@/modules/story/hooks/story";
import { useUpdateStoryMutation } from "../hooks/update-mutation";
import { useStatuses } from "@/lib/hooks/statuses";
import { useSprints } from "@/lib/hooks/sprints";
import { useMembers } from "@/lib/hooks/members";
import Link from "next/link";
import { ObjectivesMenu } from "@/components/ui/story/objectives-menu";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";

const Option = ({ label, value }: { label: string; value: ReactNode }) => {
  return (
    <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
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
    statuses.at(0))!!;
  const isDeleted = !!deletedAt;
  const assignee = members.find((m) => m.id === assigneeId);
  const reporter = members.find((m) => m.id === reporterId);

  const { mutateAsync } = useUpdateStoryMutation();

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

  const handleUpdate = async (data: Partial<DetailedStory>) => {
    await mutateAsync({ storyId: id, payload: data });
  };

  return (
    <Box className="h-full overflow-y-auto bg-gradient-to-br from-white via-gray-50/50 to-gray-50 pb-6 dark:from-dark-200/50 dark:to-dark">
      <OptionsHeader />
      <Container className="px-8 pt-4 text-gray-300/90">
        <Box className="mb-6 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text fontWeight="semibold">Properties</Text>
          {isDeleted && (
            <Badge
              size="lg"
              color="tertiary"
              className="border-opacity-30 px-2 text-dark dark:bg-opacity-30 dark:text-white"
            >
              {differenceInDays(
                addDays(new Date(deletedAt), 30),
                new Date(deletedAt),
              )}{" "}
              days left in bin
            </Badge>
          )}
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
                href={`/profile/${reporter?.id}`}
                className="relative -top-[1px] font-medium text-dark dark:text-gray-200"
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
                  disabled={isDeleted}
                  leftIcon={<StoryStatusIcon statusId={statusId} />}
                  type="button"
                  variant="naked"
                >
                  {name}
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items
                statusId={statusId}
                setStatusId={(statusId) => {
                  handleUpdate({ statusId });
                }}
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
                  disabled={isDeleted}
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
                  disabled={isDeleted}
                  color="tertiary"
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
                    <Text color="muted" as="span">
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
        {/* <Option label="Labels" value={<Labels />} /> */}
        <Option
          label="Start date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  disabled={isDeleted}
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
                    format(new Date(startDate), "MMM dd, yyyy")
                  ) : (
                    <Text color="muted">Add start date</Text>
                  )}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                selected={startDate ? new Date(startDate) : undefined}
                onDayClick={(day) => {
                  handleUpdate({
                    startDate: formatISO(day),
                  });
                }}
              />
            </DatePicker>
          }
        />
        <Option
          label="Due date"
          value={
            <DatePicker>
              <Tooltip
                hidden={!endDate}
                className="py-3"
                title={
                  <Flex align="start" gap={2}>
                    <CalendarIcon
                      className={cn("relative top-[2.5px] h-5 w-auto", {
                        "text-primary dark:text-primary":
                          new Date(endDate!!) < new Date(),
                        "text-warning dark:text-warning":
                          new Date(endDate!!) <= addDays(new Date(), 7) &&
                          new Date(endDate!!) >= new Date(),
                      })}
                    />
                    <Box>{getDueDateMessage(new Date(endDate!!))}</Box>
                  </Flex>
                }
              >
                <span>
                  <DatePicker.Trigger>
                    <Button
                      color="tertiary"
                      className={cn({
                        "text-primary dark:text-primary":
                          endDate && new Date(endDate) < new Date(),
                        "text-warning dark:text-warning":
                          endDate &&
                          new Date(endDate) <= addDays(new Date(), 7) &&
                          new Date(endDate) >= new Date(),
                        "text-gray/80 dark:text-gray-300/80": !endDate,
                      })}
                      disabled={isDeleted}
                      leftIcon={<CalendarIcon className="h-[1.15rem] w-auto" />}
                      variant="naked"
                    >
                      {endDate ? (
                        format(new Date(endDate), "MMM dd, yyyy")
                      ) : (
                        <Text color="muted">Add due date</Text>
                      )}
                    </Button>
                  </DatePicker.Trigger>
                </span>
              </Tooltip>
              <DatePicker.Calendar
                selected={endDate ? new Date(endDate) : undefined}
                onDayClick={(day) => {
                  handleUpdate({
                    endDate: formatISO(day),
                  });
                }}
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
                  title={objectiveId ? objective?.name : undefined}
                  disabled={isDeleted}
                  leftIcon={
                    objectiveId ? (
                      <ObjectiveIcon className="h-[1.15rem] w-auto" />
                    ) : (
                      <PlusIcon className="h-5 w-auto" />
                    )
                  }
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
                  disabled={isDeleted}
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
                sprintId={sprintId ?? undefined}
                setSprintId={(sprintId) => {
                  handleUpdate({ sprintId });
                }}
              />
            </SprintsMenu>
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
        <AddLinks />
      </Container>
    </Box>
  );
};
