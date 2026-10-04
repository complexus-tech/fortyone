"use client";

import type { ReactNode } from "react";
import { EstimateIcon, PlusIcon, TagsIcon, Time02Icon } from "icons";
import { cn } from "lib";
import { Avatar, Button, Flex } from "ui";
import { PriorityIcon } from "@/components/ui/priority-icon";
import { StoryStatusIcon } from "@/components/ui/story-status-icon";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { EstimateMenu } from "@/components/ui/story/estimate-menu";
import { LabelsMenu } from "@/components/ui/story/labels-menu";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { StatusesMenu } from "@/components/ui/story/statuses-menu";
import { TimeNeededMenu } from "@/components/ui/story/time-needed-menu";
import { DEFAULT_ESTIMATE_SCHEME, formatEstimate } from "@/lib/estimate";
import { useLabels } from "@/lib/hooks/labels";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { formatTimeNeeded } from "@/lib/time-needed";
import { useTeamSettings } from "@/modules/teams/public/client";
import type { AutomationDraft } from "./types";

export const RecurrenceProperties = ({
  teamId,
  draft,
  onChange,
  disabled,
  children,
}: {
  teamId: string;
  draft: AutomationDraft;
  onChange: (partial: Partial<AutomationDraft>) => void;
  disabled: boolean;
  children?: ReactNode;
}) => {
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const { data: members = [] } = useTeamMembers(teamId);
  const { data: labels = [] } = useLabels({ teamId });
  const { data: settings } = useTeamSettings(teamId);
  const estimateScheme =
    settings?.estimationSettings.scheme ?? DEFAULT_ESTIMATE_SCHEME;
  const status = statuses.find(({ id }) => id === draft.statusId);
  const member = members.find(({ id }) => id === draft.assigneeId);
  const labelIds = draft.labelIds ?? [];
  const selectedLabels = labels.filter(({ id }) => labelIds.includes(id));
  const updateDraft = (partial: Partial<AutomationDraft>) => {
    if (disabled) return;
    onChange(partial);
  };

  return (
    <Flex align="center" className="mt-4 gap-1.5" wrap>
      <StatusesMenu>
        <StatusesMenu.Trigger>
          <Button
            className="dark:bg-surface-elevated/90"
            color="tertiary"
            disabled={disabled}
            leftIcon={
              <StoryStatusIcon
                className={cn("size-4 shrink-0", {
                  "opacity-30": !draft.statusId,
                })}
                statusId={draft.statusId}
              />
            }
            size="sm"
            type="button"
            variant="outline"
          >
            {status?.name ?? "Status"}
          </Button>
        </StatusesMenu.Trigger>
        <StatusesMenu.Items
          setStatusId={(statusId) => {
            updateDraft({ statusId });
          }}
          statusId={draft.statusId ?? ""}
          teamId={teamId}
        />
      </StatusesMenu>
      <PrioritiesMenu>
        <PrioritiesMenu.Trigger>
          <Button
            className="dark:bg-surface-elevated/90"
            color="tertiary"
            disabled={disabled}
            leftIcon={
              <PriorityIcon className="h-4" priority={draft.priority} />
            }
            size="sm"
            type="button"
            variant="outline"
          >
            {draft.priority}
          </Button>
        </PrioritiesMenu.Trigger>
        <PrioritiesMenu.Items
          priority={draft.priority}
          setPriority={(priority) => {
            updateDraft({ priority });
          }}
        />
      </PrioritiesMenu>
      {selectedLabels.length ? (
        <Flex align="center" className="gap-1" wrap>
          {selectedLabels.map((label) => (
            <LabelsMenu key={label.id}>
              <LabelsMenu.Trigger>
                <Button
                  className="dark:bg-surface-elevated/90 gap-1.5 px-2.5"
                  color="tertiary"
                  disabled={disabled}
                  leftIcon={
                    <TagsIcon
                      className="h-4.5 w-auto"
                      style={{ color: label.color }}
                    />
                  }
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  <span className="inline-block max-w-[12ch] truncate">
                    {label.name}
                  </span>
                </Button>
              </LabelsMenu.Trigger>
              <LabelsMenu.Items
                labelIds={labelIds}
                setLabelIds={(nextLabelIds) => {
                  updateDraft({ labelIds: nextLabelIds });
                }}
                teamId={teamId}
              />
            </LabelsMenu>
          ))}
          <LabelsMenu>
            <LabelsMenu.Trigger>
              <Button
                asIcon
                className="dark:bg-surface-elevated/90"
                color="tertiary"
                disabled={disabled}
                leftIcon={<PlusIcon />}
                rounded="full"
                size="sm"
                title="Add labels"
                type="button"
                variant="outline"
              >
                <span className="sr-only">Add labels</span>
              </Button>
            </LabelsMenu.Trigger>
            <LabelsMenu.Items
              labelIds={labelIds}
              setLabelIds={(nextLabelIds) => {
                updateDraft({ labelIds: nextLabelIds });
              }}
              teamId={teamId}
            />
          </LabelsMenu>
        </Flex>
      ) : (
        <LabelsMenu>
          <LabelsMenu.Trigger>
            <Button
              className="dark:bg-surface-elevated/90 gap-1.5 px-2"
              color="tertiary"
              disabled={disabled}
              leftIcon={<TagsIcon className="h-4.5 w-auto" />}
              size="sm"
              type="button"
              variant="outline"
            >
              Labels
            </Button>
          </LabelsMenu.Trigger>
          <LabelsMenu.Items
            labelIds={labelIds}
            setLabelIds={(nextLabelIds) => {
              updateDraft({ labelIds: nextLabelIds });
            }}
            teamId={teamId}
          />
        </LabelsMenu>
      )}
      <AssigneesMenu>
        <AssigneesMenu.Trigger>
          <Button
            className="dark:bg-surface-elevated/90 gap-1.5 px-2"
            color="tertiary"
            disabled={disabled}
            leftIcon={
              <Avatar
                aria-hidden
                name={member?.fullName}
                size="xs"
                src={member?.avatarUrl}
              />
            }
            size="sm"
            type="button"
            variant="outline"
          >
            <span className="relative -top-px inline-block max-w-[12ch] truncate">
              {member?.username || "Assignee"}
            </span>
          </Button>
        </AssigneesMenu.Trigger>
        <AssigneesMenu.Items
          assigneeId={draft.assigneeId}
          onAssigneeSelected={(assigneeId) => {
            updateDraft({ assigneeId: assigneeId ?? undefined });
          }}
          teamId={teamId}
        />
      </AssigneesMenu>
      <EstimateMenu>
        <EstimateMenu.Trigger>
          <Button
            className={cn("dark:bg-surface-elevated/90 gap-1.5 px-2", {
              "text-text-muted": !draft.estimateValue,
            })}
            color="tertiary"
            disabled={disabled}
            leftIcon={
              <EstimateIcon
                className={cn("h-4.5 w-auto", {
                  "text-text-muted": !draft.estimateValue,
                })}
              />
            }
            size="sm"
            type="button"
            variant="outline"
          >
            {draft.estimateValue
              ? formatEstimate(estimateScheme, draft.estimateValue, "full")
              : "Complexity"}
          </Button>
        </EstimateMenu.Trigger>
        <EstimateMenu.Items
          estimateScheme={estimateScheme}
          estimateValue={draft.estimateValue}
          setEstimateValue={(estimateValue) => {
            updateDraft({ estimateValue: estimateValue ?? undefined });
          }}
        />
      </EstimateMenu>
      <TimeNeededMenu>
        <TimeNeededMenu.Trigger>
          <Button
            className={cn("dark:bg-surface-elevated/90 gap-1.5 px-2", {
              "text-text-muted": !draft.estimatedDurationMinutes,
            })}
            color="tertiary"
            disabled={disabled}
            leftIcon={
              <Time02Icon
                className={cn("h-4.5 w-auto", {
                  "text-text-muted": !draft.estimatedDurationMinutes,
                })}
              />
            }
            size="sm"
            type="button"
            variant="outline"
          >
            {draft.estimatedDurationMinutes
              ? formatTimeNeeded(draft.estimatedDurationMinutes, "full")
              : "Time needed"}
          </Button>
        </TimeNeededMenu.Trigger>
        <TimeNeededMenu.Items
          estimatedDurationMinutes={draft.estimatedDurationMinutes}
          minimumFocusBlockMinutes={draft.minimumFocusBlockMinutes}
          setTimeNeeded={({
            estimatedDurationMinutes,
            minimumFocusBlockMinutes,
          }) => {
            updateDraft({
              estimatedDurationMinutes: estimatedDurationMinutes ?? undefined,
              minimumFocusBlockMinutes: minimumFocusBlockMinutes ?? undefined,
            });
          }}
        />
      </TimeNeededMenu>
      {children}
    </Flex>
  );
};
