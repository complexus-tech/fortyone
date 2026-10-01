"use client";

import { Text } from "ui";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useTeamMembers } from "@/lib/hooks/team-members";
import type { Automation } from "./types";

export const AutomationDetails = ({
  automation,
}: {
  automation: Automation;
}) => {
  const statuses = useTeamStatuses(automation.teamId);
  const members = useTeamMembers(automation.teamId);
  if (automation.kind === "recurrence") {
    const { schedule, draft } = automation.configuration;
    return (
      <Text className="mt-2" color="muted">
        Creates “{draft.title}” {schedule.frequency} at {schedule.localTime} (
        {schedule.timezone}).
      </Text>
    );
  }
  const { trigger, conditions, actions } = automation.configuration;
  const statusName = (id: string) =>
    statuses.data?.find((status) => status.id === id)?.name ??
    "Unavailable status";
  const memberName = (id: string) =>
    members.data?.find((member) => member.id === id)?.fullName ??
    "Unavailable member";
  const conditionLabels = [
    conditions.statusIds?.length
      ? `status is ${conditions.statusIds.map(statusName).join(" or ")}`
      : null,
    conditions.priorities?.length
      ? `priority is ${conditions.priorities.join(" or ")}`
      : null,
    conditions.assigneeIds?.length
      ? `assigned to ${conditions.assigneeIds.map(memberName).join(" or ")}`
      : null,
    conditions.unassigned ? "unassigned" : null,
  ].filter(Boolean);
  const actionLabels = [
    actions.statusId ? `set status to ${statusName(actions.statusId)}` : null,
    actions.priority ? `set priority to ${actions.priority}` : null,
    actions.assigneeId ? `assign to ${memberName(actions.assigneeId)}` : null,
    actions.clearAssignee ? "remove assignee" : null,
  ].filter(Boolean);
  return (
    <Text className="mt-2" color="muted">
      When a task is {trigger === "story.created" ? "created" : "updated"}
      {conditionLabels.length
        ? ` and ${conditionLabels.join(" and ")}`
        : ""}, {actionLabels.join(" and ")}.
    </Text>
  );
};
