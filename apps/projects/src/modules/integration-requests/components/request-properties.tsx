"use client";

import { format, formatISO } from "date-fns";
import { cn } from "lib";
import {
  CalendarIcon,
  CloseIcon,
  EstimateIcon,
  ObjectiveIcon,
  SprintsIcon,
  TagsIcon,
  Time02Icon,
} from "icons";
import { Avatar, Box, Button, Container, DatePicker, Text } from "ui";
import { ObjectiveKeyResultMenu } from "@/components/ui/story/objective-key-result-menu";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { EstimateMenu } from "@/components/ui/story/estimate-menu";
import { LabelsMenu } from "@/components/ui/story/labels-menu";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { SprintsMenu } from "@/components/ui/story/sprints-menu";
import { StatusesMenu } from "@/components/ui/story/statuses-menu";
import { TimeNeededMenu } from "@/components/ui/story/time-needed-menu";
import { PriorityIcon } from "@/components/ui/priority-icon";
import { PropertyOption } from "@/components/ui/property-option";
import { StoryStatusIcon } from "@/components/ui/story-status-icon";
import { useTerminology } from "@/hooks/use-terminology-display";
import {
  DEFAULT_ESTIMATE_SCHEME,
  formatEstimate,
  type EstimateScheme,
} from "@/lib/estimate";
import { useLabels } from "@/lib/hooks/labels";
import { formatTimeNeeded } from "@/lib/time-needed";
import {
  useKeyResults,
  useObjective,
} from "@/modules/objectives/public/client";
import { useSprint } from "@/modules/sprints/public/client";
import { useTeamSettings } from "@/modules/teams/public/client";
import type { State } from "@/types/states";
import type {
  IntegrationRequest,
  UpdateIntegrationRequestInput,
} from "../types";

type RequestAssignee = {
  avatarUrl: string | null;
  fullName: string;
  username: string;
};

type RequestPropertiesProps = {
  assignee?: RequestAssignee;
  canEditRequest: boolean;
  onUpdate: (payload: UpdateIntegrationRequestInput) => void;
  priority: IntegrationRequest["priority"];
  request: IntegrationRequest;
  selectedStatus?: State;
  statusId?: string;
  teamId: string;
  variant?: "sidebar" | "inline";
};

export const RequestProperties = ({
  assignee,
  canEditRequest,
  onUpdate,
  priority,
  request,
  selectedStatus,
  statusId,
  teamId,
  variant = "sidebar",
}: RequestPropertiesProps) => {
  const isInline = variant === "inline";
  const { getTermDisplay } = useTerminology();
  const { data: teamSettings } = useTeamSettings(teamId);
  const { data: selectedObjective } = useObjective(
    request.objectiveId ?? null,
    teamId,
  );
  const { data: keyResults = [] } = useKeyResults(
    request.objectiveId ?? "",
    Boolean(request.objectiveId),
  );
  const selectedKeyResult = keyResults.find(
    (keyResult) => keyResult.id === request.keyResultId,
  );
  const { data: allLabels = [] } = useLabels({ teamId });
  const selectedLabelIds = new Set(request.labelIds);
  const selectedLabels = allLabels.filter((label) =>
    selectedLabelIds.has(label.id),
  );
  const { data: selectedSprint } = useSprint(request.sprintId ?? null, teamId);
  const estimateScheme = (teamSettings?.estimationSettings.scheme ??
    DEFAULT_ESTIMATE_SCHEME) as EstimateScheme;
  const requestEstimateLabel = formatEstimate(
    estimateScheme,
    request.estimateValue,
    "compact",
  );

  return (
    <Container
      className={cn("text-text-muted px-0.5 pt-4 md:px-6", {
        "px-0 pt-0 md:px-0": isInline,
      })}
    >
      {!isInline ? (
        <Box className="mb-0 grid grid-cols-[9rem_auto] items-center gap-3 md:mb-6">
          <Text className="hidden md:block" fontWeight="semibold">
            Properties
          </Text>
        </Box>
      ) : null}

      <Box className={cn("flex flex-wrap gap-2", { "md:block": !isInline })}>
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label="Status"
          value={
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<StoryStatusIcon statusId={statusId} />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {selectedStatus?.name ?? "Todo"}
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items
                setStatusId={(nextStatusId) => {
                  onUpdate({ statusId: nextStatusId });
                }}
                statusId={statusId}
                teamId={teamId}
              />
            </StatusesMenu>
          }
        />
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label="Priority"
          value={
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<PriorityIcon priority={priority} />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {priority}
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items
                priority={priority}
                setPriority={(nextPriority: IntegrationRequest["priority"]) => {
                  onUpdate({ priority: nextPriority });
                }}
              />
            </PrioritiesMenu>
          }
        />
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label="Assignee"
          value={
            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <Button
                  className="font-medium"
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={
                    <Avatar
                      className="text-foreground/80"
                      name={assignee?.fullName}
                      size="xs"
                      src={assignee?.avatarUrl}
                    />
                  }
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {assignee?.username ?? (
                    <Text as="span" color="muted">
                      Assign
                    </Text>
                  )}
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items
                assigneeId={request.assigneeId}
                onAssigneeSelected={(assigneeId) => {
                  onUpdate({ assigneeId: assigneeId ?? null });
                }}
                teamId={teamId}
              />
            </AssigneesMenu>
          }
        />
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label="Complexity"
          value={
            <EstimateMenu>
              <EstimateMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<EstimateIcon className="h-4" />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {requestEstimateLabel}
                </Button>
              </EstimateMenu.Trigger>
              <EstimateMenu.Items
                align="start"
                estimateScheme={estimateScheme}
                estimateValue={request.estimateValue}
                setEstimateValue={(estimateValue) => {
                  onUpdate({ estimateValue });
                }}
              />
            </EstimateMenu>
          }
        />
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label="Time needed"
          value={
            <TimeNeededMenu>
              <TimeNeededMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<Time02Icon className="h-4" />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {formatTimeNeeded(
                    request.estimatedDurationMinutes,
                    request.estimatedDurationMinutes ? "full" : "compact",
                  )}
                </Button>
              </TimeNeededMenu.Trigger>
              <TimeNeededMenu.Items
                align="start"
                estimatedDurationMinutes={request.estimatedDurationMinutes}
                minimumFocusBlockMinutes={request.minimumFocusBlockMinutes}
                setTimeNeeded={(timeNeeded) => {
                  onUpdate(timeNeeded);
                }}
              />
            </TimeNeededMenu>
          }
        />
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label={getTermDisplay("objectiveTerm", { capitalize: true })}
          value={
            <ObjectiveKeyResultMenu
              keyResultId={request.keyResultId ?? null}
              objectiveId={request.objectiveId ?? null}
              onChange={({ keyResultId, objectiveId }) => {
                onUpdate({
                  objectiveId: objectiveId ?? null,
                  keyResultId: keyResultId ?? null,
                });
              }}
              teamId={teamId}
            >
              <Button
                color="tertiary"
                disabled={!canEditRequest}
                leftIcon={<ObjectiveIcon className="h-4" />}
                size="sm"
                variant={isInline ? "solid" : "naked"}
              >
                <span className="max-w-40 truncate">
                  {selectedKeyResult?.name ??
                    selectedObjective?.name ??
                    getTermDisplay("objectiveTerm", { capitalize: true })}
                </span>
              </Button>
            </ObjectiveKeyResultMenu>
          }
        />
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label="Labels"
          value={
            <LabelsMenu>
              <LabelsMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<TagsIcon className="h-4" />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  <span className="max-w-40 truncate">
                    {selectedLabels.length > 0
                      ? selectedLabels.map((label) => label.name).join(", ")
                      : "Add labels"}
                  </span>
                </Button>
              </LabelsMenu.Trigger>
              <LabelsMenu.Items
                labelIds={request.labelIds}
                setLabelIds={(labelIds) => {
                  onUpdate({ labelIds });
                }}
                teamId={teamId}
              />
            </LabelsMenu>
          }
        />
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label={getTermDisplay("sprintTerm", { capitalize: true })}
          value={
            <SprintsMenu>
              <SprintsMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<SprintsIcon className="h-[1.05rem]" />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  <span className="max-w-40 truncate">
                    {selectedSprint?.name ??
                      getTermDisplay("sprintTerm", { capitalize: true })}
                  </span>
                </Button>
              </SprintsMenu.Trigger>
              <SprintsMenu.Items
                setSprintId={(sprintId) => {
                  onUpdate({ sprintId: sprintId ?? null });
                }}
                sprintId={request.sprintId}
                teamId={teamId}
              />
            </SprintsMenu>
          }
        />
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label="Start"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<CalendarIcon className="h-4" />}
                  rightIcon={
                    request.startDate ? (
                      <CloseIcon
                        aria-label="Remove start date"
                        className="h-4"
                        onClick={(event) => {
                          event.preventDefault();
                          event.stopPropagation();
                          onUpdate({ startDate: null });
                        }}
                        onPointerDown={(event) => {
                          event.preventDefault();
                          event.stopPropagation();
                        }}
                        role="button"
                      />
                    ) : null
                  }
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {request.startDate
                    ? format(new Date(request.startDate), "MMM d")
                    : "Start"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day) => {
                  onUpdate({
                    startDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  request.startDate ? new Date(request.startDate) : undefined
                }
              />
            </DatePicker>
          }
        />
        <PropertyOption
          isCompact={isInline}
          isNotifications={isInline}
          label="Deadline"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<CalendarIcon className="h-4" />}
                  rightIcon={
                    request.endDate ? (
                      <CloseIcon
                        aria-label="Remove deadline"
                        className="h-4"
                        onClick={(event) => {
                          event.preventDefault();
                          event.stopPropagation();
                          onUpdate({ endDate: null });
                        }}
                        onPointerDown={(event) => {
                          event.preventDefault();
                          event.stopPropagation();
                        }}
                        role="button"
                      />
                    ) : null
                  }
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {request.endDate
                    ? format(new Date(request.endDate), "MMM d")
                    : "Deadline"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day) => {
                  onUpdate({
                    endDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  request.endDate ? new Date(request.endDate) : undefined
                }
              />
            </DatePicker>
          }
        />
      </Box>
    </Container>
  );
};
