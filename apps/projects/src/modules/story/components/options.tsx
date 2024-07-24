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
import { useState, type ReactNode } from "react";
import { addDays, format, differenceInDays, isTomorrow } from "date-fns";
import type { DateRange } from "react-day-picker";
import { CalendarIcon } from "icons";
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
import { useStore } from "@/hooks/store";
import { cn } from "lib";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { toast } from "sonner";

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

export const Options = ({ story }: { story: DetailedStory }) => {
  const {
    id,
    priority,
    statusId,
    startDate,
    endDate,
    objectiveId,
    sprintId,
    deletedAt,
  } = story;
  const { states } = useStore();
  const { name } = (states.find((state) => state.id === statusId) ||
    states.at(0))!!;
  const isDeleted = !!deletedAt;
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
    try {
      await updateStoryAction(id, data);
    } catch (error) {
      console.log(error);
      toast.error("Error updating story", {
        description: "Failed to update story, reverted changes.",
      });
    }
  };

  return (
    <Box className="h-full overflow-y-auto bg-gradient-to-br from-white via-gray-50/50 to-gray-50 pb-6 dark:from-dark-200/50 dark:to-dark">
      <OptionsHeader story={story} />
      <Container className="px-8 pt-4 text-gray-300/90">
        <Flex className="mb-5" align="center" justify="between">
          <Text fontWeight="semibold">Properties</Text>
          {isDeleted && (
            <Badge
              rounded="full"
              color="warning"
              className="text-dark dark:bg-opacity-50 dark:text-white"
            >
              {differenceInDays(
                addDays(new Date(deletedAt), 30),
                new Date(deletedAt),
              )}{" "}
              days left in bin
            </Badge>
          )}
        </Flex>
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
                setStatusId={(st) => {
                  handleUpdate({ statusId: st });
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
                setPriority={(p) => {
                  handleUpdate({ priority: p });
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
                  className="font-medium"
                  disabled={isDeleted}
                  color="tertiary"
                  leftIcon={
                    <Avatar
                      name="Joseph Mukorivo"
                      size="xs"
                      src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                    />
                  }
                  type="button"
                  variant="naked"
                >
                  josemukorivo
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items />
            </AssigneesMenu>
          }
        />
        <Option label="Labels" value={<Labels />} />
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
                  handleUpdate({ startDate: day.toISOString() });
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
                  handleUpdate({ endDate: day.toISOString() });
                }}
              />
            </DatePicker>
          }
        />
        <Option label="Sprint" value={<SprintsMenu />} />
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
        />

        <Divider className="my-4" />
        <AddLinks />
      </Container>
    </Box>
  );
};
