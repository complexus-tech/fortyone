import type { Objective } from "./types";
import type { ObjectivePatch } from "./detail-data";
import { useState } from "react";
import { formatISO } from "date-fns";
import { Ionicons } from "@expo/vector-icons";
import { StatusIcon } from "@/components/icons";
import { DateField } from "@/components/ui/date-field";
import {
  DetailProperties,
  SelectProperty,
} from "@/modules/entity-details/properties";
import { PriorityBadge } from "@/modules/story/components/properties/priority";
import { AssigneeBadge } from "@/modules/story/components/properties/assignee";
import { PropertyExpandButton } from "@/modules/story/components/properties/property-expand-button";
import { useObjectiveStatuses } from "./hooks/use-objectives";
import { colors } from "@/constants/colors";

export function ObjectiveProperties({
  item,
  disabled,
  onUpdate,
}: {
  item: Objective;
  disabled: boolean;
  onUpdate: (patch: ObjectivePatch) => Promise<void>;
}) {
  const [expanded, setExpanded] = useState(false);
  const statuses = useObjectiveStatuses();
  const status = statuses.data?.find((s) => s.id === item.statusId);
  const healthOptions = ["On Track", "At Risk", "Off Track"] as const;
  return (
    <DetailProperties>
      <SelectProperty
        title="Status"
        label={status?.name ?? "Status"}
        icon={<StatusIcon category={status?.category} color={status?.color} />}
        disabled={disabled}
        loading={statuses.isPending}
        error={statuses.error}
        onRetry={() => void statuses.refetch()}
        selectedIds={[item.statusId]}
        options={(statuses.data ?? []).map((s) => ({
          id: s.id,
          label: s.name,
          icon: <StatusIcon category={s.category} color={s.color} />,
        }))}
        onSelect={(statusId) => onUpdate({ statusId })}
      />
      <PriorityBadge
        priority={item.priority ?? "No Priority"}
        disabled={disabled}
        onPriorityChange={(priority) => onUpdate({ priority })}
      />
      <AssigneeBadge
        story={{ assigneeId: item.leadUser }}
        disabled={disabled}
        onAssigneeChange={(leadUser) => onUpdate({ leadUser })}
        label="Lead"
      />
      {expanded || item.health ? (
        <SelectProperty
          title="Health"
          label={item.health ?? "Health"}
          disabled={disabled}
          icon={
            <Ionicons
              name="pulse-outline"
              size={18}
              color={
                item.health === "On Track"
                  ? colors.success
                  : item.health === "Off Track"
                    ? colors.danger
                    : colors.warning
              }
            />
          }
          options={healthOptions.map((health) => ({
            id: health,
            label: health,
          }))}
          selectedIds={item.health ? [item.health] : []}
          onSelect={async (value) => {
            const health = healthOptions.find((h) => h === value);
            if (health) await onUpdate({ health });
          }}
          clearLabel="No health"
          onClear={() => onUpdate({ health: null })}
        />
      ) : null}
      {expanded || item.startDate ? (
        <DateField
          label="Start date"
          value={item.startDate}
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
          value={item.endDate}
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
