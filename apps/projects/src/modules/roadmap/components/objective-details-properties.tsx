"use client";

import type { ReactNode } from "react";
import { format, formatISO } from "date-fns";
import { CalendarIcon, ClockIcon, WarningIcon } from "icons";
import { cn } from "lib";
import { Avatar, Box, Button, ColorPicker, DatePicker, Flex, Text } from "ui";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useCanUpdateObjective } from "@/modules/objectives/hooks";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks/update-mutation";
import type { Objective, ObjectiveUpdate } from "@/modules/objectives/types";
import { ObjectiveHealthEditor } from "@/modules/objectives/components/objective-health-editor";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { ObjectiveHealthIcon } from "@/components/ui/objective-health-icon";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { PriorityIcon } from "@/components/ui/priority-icon";
import { ObjectivePillarProperty } from "@/modules/strategy/objective-pillar-property";

const DetailRow = ({ label, value }: { label: string; value: ReactNode }) => (
  <Flex align="center" className="min-h-9" gap={4}>
    <Text className="w-28 shrink-0" color="muted">
      {label}
    </Text>
    <Box className="min-w-0 flex-1">{value}</Box>
  </Flex>
);

const propertyButtonClassName =
  "-ml-2 min-w-0 max-w-full justify-start px-2 font-normal";

export const ObjectiveDetailsProperties = ({
  objective,
}: {
  objective: Objective;
}) => {
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: members = [] } = useTeamMembers(objective.teamId);
  const canUpdate = useCanUpdateObjective();
  const updateMutation = useUpdateObjectiveMutation();
  const status = statuses.find((item) => item.id === objective.statusId);
  const lead = members.find((member) => member.id === objective.leadUser);

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId: objective.id,
      data,
    });
  };

  return (
    <>
      <Text className="mb-3">Properties</Text>
      <Flex direction="column" gap={2}>
        <ObjectivePillarProperty
          buttonClassName={propertyButtonClassName}
          layout="detail"
          objectiveId={objective.id}
        />
        <DetailRow
          label="Status"
          value={
            <ObjectiveStatusesMenu>
              <ObjectiveStatusesMenu.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <ObjectiveStatusIcon statusId={objective.statusId} />
                  }
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  <span className="truncate">
                    {status?.name ?? "No status"}
                  </span>
                </Button>
              </ObjectiveStatusesMenu.Trigger>
              <ObjectiveStatusesMenu.Items
                setStatusId={(statusId) => {
                  handleUpdate({ statusId });
                }}
                statusId={objective.statusId}
              />
            </ObjectiveStatusesMenu>
          }
        />
        <DetailRow
          label="Health"
          value={
            <ObjectiveHealthEditor
              health={objective.health}
              objectiveId={objective.id}
            >
              <Button
                align="left"
                className={propertyButtonClassName}
                color="tertiary"
                disabled={!canUpdate}
                leftIcon={<ObjectiveHealthIcon health={objective.health} />}
                size="sm"
                type="button"
                variant="naked"
              >
                {objective.health ?? "No health"}
              </Button>
            </ObjectiveHealthEditor>
          }
        />
        <DetailRow
          label="Color"
          value={
            <Flex align="center" gap={1}>
              <ColorPicker
                className={cn("-ml-2", {
                  "pointer-events-none opacity-60": !canUpdate,
                })}
                onChange={(color) => {
                  handleUpdate({ color });
                }}
                value={objective.color}
              />
              <Text color="muted" fontSize="sm">
                {objective.color.toUpperCase()}
              </Text>
            </Flex>
          }
        />
        <DetailRow
          label="Priority"
          value={
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={<PriorityIcon priority={objective.priority} />}
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  {objective.priority ?? "No priority"}
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items
                priority={objective.priority}
                setPriority={(priority) => {
                  handleUpdate({ priority });
                }}
              />
            </PrioritiesMenu>
          }
        />
        <DetailRow
          label="Lead"
          value={
            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <Avatar
                      className={cn({
                        "text-foreground/80": !objective.leadUser,
                      })}
                      name={lead?.fullName || lead?.username}
                      size="xs"
                      src={lead?.avatarUrl}
                    />
                  }
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  <span className="truncate">
                    {lead?.username ?? "Assign lead"}
                  </span>
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items
                assigneeId={objective.leadUser}
                onAssigneeSelected={(leadUser) => {
                  handleUpdate({ leadUser: leadUser ?? undefined });
                }}
                placeholder="Assign lead..."
                teamId={objective.teamId}
              />
            </AssigneesMenu>
          }
        />
        <DetailRow
          label="Start date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <CalendarIcon
                      className={cn("h-4 w-auto", {
                        "text-text-muted": !objective.startDate,
                      })}
                    />
                  }
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  {objective.startDate
                    ? format(new Date(objective.startDate), "MMM d, yyyy")
                    : "Start date"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day) => {
                  handleUpdate({
                    startDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  objective.startDate
                    ? new Date(objective.startDate)
                    : undefined
                }
              />
            </DatePicker>
          }
        />
        <DetailRow
          label="Target date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <CalendarIcon
                      className={cn("h-4 w-auto", {
                        "text-text-muted": !objective.endDate,
                      })}
                    />
                  }
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  {objective.endDate
                    ? format(new Date(objective.endDate), "MMM d, yyyy")
                    : "Target date"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day) => {
                  handleUpdate({
                    endDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  objective.endDate ? new Date(objective.endDate) : undefined
                }
              />
            </DatePicker>
          }
        />
        {objective.forecastEndDate ? (
          <DetailRow
            label="Forecast"
            value={
              <Box className="py-1">
                <Flex align="center" gap={2}>
                  {objective.scheduleStatus === "at_risk" ? (
                    <WarningIcon className="text-danger h-4 w-auto shrink-0" />
                  ) : (
                    <ClockIcon className="text-text-muted h-4 w-auto shrink-0" />
                  )}
                  <Text
                    color={
                      objective.scheduleStatus === "at_risk"
                        ? "danger"
                        : undefined
                    }
                    fontWeight="medium"
                  >
                    {format(new Date(objective.forecastEndDate), "MMM d, yyyy")}
                    {objective.forecastDaysDelta > 0
                      ? ` · ${objective.forecastDaysDelta} days late`
                      : " · On track"}
                  </Text>
                </Flex>
                {objective.forecastCauseStory ? (
                  <Text className="mt-0.5 pl-6" color="muted" fontSize="sm">
                    Driven by task {objective.forecastCauseStory.sequenceId}:{" "}
                    {objective.forecastCauseStory.title}
                  </Text>
                ) : null}
              </Box>
            }
          />
        ) : null}
      </Flex>
    </>
  );
};
