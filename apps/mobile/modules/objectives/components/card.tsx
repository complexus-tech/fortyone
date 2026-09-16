import type { Objective, ObjectiveStatus } from "../types";
import { memo } from "react";
import { useRouter } from "expo-router";
import { useTerminology } from "@/hooks/use-terminology";
import { PriorityIcon, StatusIcon } from "@/components/icons";
import { WorkItemRow } from "@/modules/teams/stories/components/work-item-row";

export const Card = memo(function Card({
  objective,
  status,
}: {
  objective: Objective;
  status?: ObjectiveStatus;
}) {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  return (
    <WorkItemRow
      title={objective.name}
      leading={
        <>
          <StatusIcon
            size={20}
            category={status?.category}
            color={status?.color}
          />
          <PriorityIcon
            size={18}
            priority={objective.priority ?? "No Priority"}
          />
        </>
      }
      accessibilityLabel={[objective.name, status?.name, objective.priority]
        .filter(Boolean)
        .join(", ")}
      accessibilityHint={`View ${getTermDisplay("objectiveTerm")} details`}
      onPress={() =>
        router.push(`/team/${objective.teamId}/objectives/${objective.id}`)
      }
    />
  );
});
