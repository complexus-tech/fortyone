import type { IntakeItem } from "./types";
import type { IntakePatch } from "./detail-data";
import { useState } from "react";
import { formatISO } from "date-fns";
import { DateField } from "@/components/ui/date-field";
import { useFeatures, useSprintsEnabled } from "@/hooks";
import { DetailProperties } from "@/modules/entity-details/properties";
import { StatusBadge } from "@/modules/story/components/properties/status";
import { PriorityBadge } from "@/modules/story/components/properties/priority";
import { AssigneeBadge } from "@/modules/story/components/properties/assignee";
import { ObjectiveBadge } from "@/modules/story/components/properties/objective";
import { SprintBadge } from "@/modules/story/components/properties/sprint";
import { LabelsBadge } from "@/modules/story/components/properties/labels";
import { PropertyExpandButton } from "@/modules/story/components/properties/property-expand-button";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";

export function IntakeProperties({
  item,
  disabled,
  onUpdate,
}: {
  item: IntakeItem;
  disabled: boolean;
  onUpdate: (patch: IntakePatch) => Promise<void>;
}) {
  const [expanded, setExpanded] = useState(false);
  const { objectiveEnabled } = useFeatures();
  const sprintsEnabled = useSprintsEnabled(item.teamId);
  const { data: statuses = [] } = useTeamStatuses(item.teamId);
  const defaultStatus =
    statuses.find((status) => status.category === "unstarted") ?? statuses[0];
  return (
    <DetailProperties>
      <StatusBadge
        story={{
          teamId: item.teamId,
          statusId: item.statusId ?? defaultStatus?.id ?? "",
        }}
        disabled={disabled}
        onStatusChange={(statusId) => onUpdate({ statusId })}
      />
      <PriorityBadge
        priority={item.priority ?? "No Priority"}
        disabled={disabled}
        onPriorityChange={(priority) => onUpdate({ priority })}
      />
      <AssigneeBadge
        story={{ assigneeId: item.assigneeId ?? null }}
        disabled={disabled}
        onAssigneeChange={(assigneeId) => onUpdate({ assigneeId })}
      />
      {expanded || item.labelIds?.length ? (
        <LabelsBadge
          story={{ teamId: item.teamId, labels: item.labelIds ?? [] }}
          disabled={disabled}
          onLabelsChange={(labelIds) => onUpdate({ labelIds })}
        />
      ) : null}
      {objectiveEnabled && (expanded || item.objectiveId) ? (
        <ObjectiveBadge
          story={{ teamId: item.teamId, objectiveId: item.objectiveId ?? null }}
          disabled={disabled}
          onObjectiveChange={(objectiveId) => onUpdate({ objectiveId })}
        />
      ) : null}
      {sprintsEnabled && (expanded || item.sprintId) ? (
        <SprintBadge
          story={{ teamId: item.teamId, sprintId: item.sprintId ?? null }}
          disabled={disabled}
          onSprintChange={(sprintId) => onUpdate({ sprintId })}
        />
      ) : null}
      {expanded || item.startDate ? (
        <DateField
          label="Start date"
          value={item.startDate ?? null}
          disabled={disabled}
          onChange={(date) =>
            onUpdate({
              startDate: date
                ? formatISO(date, { representation: "date" })
                : null,
            })
          }
        />
      ) : null}
      {expanded || item.endDate ? (
        <DateField
          label="Deadline"
          value={item.endDate ?? null}
          disabled={disabled}
          onChange={(date) =>
            onUpdate({
              endDate: date
                ? formatISO(date, { representation: "date" })
                : null,
            })
          }
        />
      ) : null}
      <PropertyExpandButton
        expanded={expanded}
        onPress={() => setExpanded(!expanded)}
      />
    </DetailProperties>
  );
}
