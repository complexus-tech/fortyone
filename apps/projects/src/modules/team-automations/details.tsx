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
    <dl className="text-text-muted mt-3 grid grid-cols-[3rem_minmax(0,1fr)] gap-x-3 gap-y-1">
      <dt>When</dt>
      <dd>A task is {trigger === "story.created" ? "created" : "updated"}</dd>
      {conditionLabels.length ? (
        <>
          <dt>If</dt>
          <dd>{conditionLabels.join(" and ")}</dd>
        </>
      ) : null}
      <dt>Then</dt>
      <dd>{actionLabels.join(" and ")}</dd>
    </dl>
  );
};
