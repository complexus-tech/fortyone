"use client";

import { Avatar, Button, Text } from "ui";
import { EstimateIcon, Time02Icon, TimeScheduleIcon } from "icons";
import { cn } from "lib";
import { PropertyOption as Option } from "@/components/ui/property-option";
import { PriorityIcon } from "@/components/ui/priority-icon";
import { StoryStatusIcon } from "@/components/ui/story-status-icon";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { AutoSchedulingMenu } from "@/components/ui/story/auto-scheduling-menu";
import { EstimateMenu } from "@/components/ui/story/estimate-menu";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { StatusesMenu } from "@/components/ui/story/statuses-menu";
import { TimeNeededMenu } from "@/components/ui/story/time-needed-menu";
import { isMayaAssigneeSelection } from "@/lib/auto-scheduling";
import { formatEstimate } from "@/lib/estimate";
import { formatTimeNeeded } from "@/lib/time-needed";
import type { DetailedStory } from "../../types";
import type { CollaboratorSummary } from "./collaborator-selection";
import { CollaboratorsOption } from "./collaborators-option";
import { useOptionHotkey } from "./use-option-hotkey";

type CoreOptionStory = Omit<
  Pick<
    DetailedStory,
    | "assigneeId"
    | "autoSchedulingEnabled"
    | "autoSchedulingLocked"
    | "collaboratorIds"
    | "collaborators"
    | "estimateScheme"
    | "estimateValue"
    | "estimatedDurationMinutes"
    | "minimumFocusBlockMinutes"
    | "priority"
    | "statusId"
    | "teamId"
  >,
  "collaboratorIds"
> & {
  collaboratorIds?: DetailedStory["collaboratorIds"] | null;
};

type CoreOptionsProps = {
  assignee?: CollaboratorSummary | null;
  canUseBackgroundMaya: boolean;
  disabled: boolean;
  isCompact: boolean;
  isNotifications: boolean;
  mayaAssigneeId?: string;
  members: CollaboratorSummary[];
  onUpdate: (story: Partial<DetailedStory>) => void;
  statusName?: string;
  story: CoreOptionStory;
  storyId: string;
};

export const CoreOptions = ({
  assignee,
  canUseBackgroundMaya,
  disabled,
  isCompact,
  isNotifications,
  mayaAssigneeId,
  members,
  onUpdate,
  statusName,
  story,
  storyId,
}: CoreOptionsProps) => {
  const {
    assigneeId,
    autoSchedulingEnabled,
    autoSchedulingLocked,
    collaboratorIds: persistedCollaboratorIds,
    collaborators,
    estimateScheme,
    estimateValue,
    estimatedDurationMinutes,
    minimumFocusBlockMinutes,
    priority,
    statusId,
    teamId,
  } = story;
  const collaboratorIds = persistedCollaboratorIds ?? [];
  const statusButtonRef = useOptionHotkey("s", !disabled);
  const priorityButtonRef = useOptionHotkey("p", !disabled);
  const assigneeButtonRef = useOptionHotkey("a", !disabled);
  const estimateButtonRef = useOptionHotkey("e", !disabled);

  return (
    <>
      <Option
        isCompact={isCompact}
        isNotifications={isNotifications}
        label="Status"
        value={
          <StatusesMenu>
            <StatusesMenu.Trigger>
              <Button
                color="tertiary"
                disabled={disabled}
                leftIcon={<StoryStatusIcon statusId={statusId} />}
                ref={statusButtonRef}
                size="sm"
                type="button"
                variant={isCompact ? "solid" : "naked"}
              >
                {statusName}
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
      <Option
        isCompact={isCompact}
        isNotifications={isNotifications}
        label="Priority"
        value={
          <PrioritiesMenu>
            <PrioritiesMenu.Trigger>
              <Button
                color="tertiary"
                disabled={disabled}
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
              setPriority={(nextPriority) => {
                onUpdate({ priority: nextPriority });
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
                disabled={disabled}
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
              onAssigneeSelected={(nextAssigneeId) => {
                onUpdate({
                  assigneeId: nextAssigneeId,
                  ...(isMayaAssigneeSelection(nextAssigneeId, mayaAssigneeId)
                    ? { autoSchedulingEnabled: true }
                    : {}),
                });
              }}
              teamId={teamId}
            />
          </AssigneesMenu>
        }
      />
      <CollaboratorsOption
        assigneeId={assigneeId}
        collaboratorIds={collaboratorIds}
        collaborators={collaborators}
        disabled={disabled}
        isCompact={isCompact}
        isNotifications={isNotifications}
        members={members}
        storyId={storyId}
        teamId={teamId}
      />
      <Option
        isCompact={isCompact}
        isNotifications={isNotifications}
        label="Complexity"
        value={
          <EstimateMenu>
            <EstimateMenu.Trigger>
              <Button
                className={cn("font-medium", {
                  "text-text-muted": !estimateValue,
                })}
                color="tertiary"
                disabled={disabled}
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
                    Add complexity
                  </Text>
                )}
              </Button>
            </EstimateMenu.Trigger>
            <EstimateMenu.Items
              estimateScheme={estimateScheme}
              estimateValue={estimateValue}
              setEstimateValue={(nextEstimateValue) => {
                onUpdate({ estimateValue: nextEstimateValue });
              }}
            />
          </EstimateMenu>
        }
      />
      <Option
        isCompact={isCompact}
        isNotifications={isNotifications}
        label="Time needed"
        value={
          <TimeNeededMenu>
            <TimeNeededMenu.Trigger>
              <Button
                className={cn("font-medium", {
                  "text-text-muted": !estimatedDurationMinutes,
                })}
                color="tertiary"
                disabled={disabled}
                leftIcon={
                  <Time02Icon
                    className={cn("h-[1.15rem] w-auto", {
                      "text-text-muted": !estimatedDurationMinutes,
                    })}
                  />
                }
                size="sm"
                type="button"
                variant={isCompact ? "solid" : "naked"}
              >
                {estimatedDurationMinutes ? (
                  formatTimeNeeded(estimatedDurationMinutes, "full")
                ) : (
                  <Text as="span" color="muted">
                    Add time needed
                  </Text>
                )}
              </Button>
            </TimeNeededMenu.Trigger>
            <TimeNeededMenu.Items
              estimatedDurationMinutes={estimatedDurationMinutes}
              minimumFocusBlockMinutes={minimumFocusBlockMinutes}
              setTimeNeeded={onUpdate}
            />
          </TimeNeededMenu>
        }
      />
      <Option
        isCompact={isCompact}
        isNotifications={isNotifications}
        label="Auto-scheduling"
        value={
          <AutoSchedulingMenu>
            <AutoSchedulingMenu.Trigger>
              <Button
                className="font-medium"
                color="tertiary"
                disabled={disabled || !canUseBackgroundMaya}
                leftIcon={<TimeScheduleIcon className="h-[1.15rem] w-auto" />}
                size="sm"
                type="button"
                variant={isCompact ? "solid" : "naked"}
              >
                {autoSchedulingEnabled ? "On" : "Off"}
              </Button>
            </AutoSchedulingMenu.Trigger>
            <AutoSchedulingMenu.Items
              autoSchedulingEnabled={autoSchedulingEnabled}
              autoSchedulingLocked={autoSchedulingLocked}
              setAutoSchedulingEnabled={(enabled) => {
                onUpdate({ autoSchedulingEnabled: enabled });
              }}
            />
          </AutoSchedulingMenu>
        }
      />
    </>
  );
};
