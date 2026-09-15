import type { Objective } from "../types";
import { useRouter } from "expo-router";
import { useTerminology } from "@/hooks/use-terminology";
import { CollectionRow } from "@/modules/teams/stories/components/collection-row";
import { formatContextDates } from "@/modules/teams/stories/context-dates";

export const Card = ({ objective }: { objective: Objective }) => {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  const context = objective.health;
  const subtitle = [
    formatContextDates(objective.startDate, objective.endDate),
    context,
  ]
    .filter(Boolean)
    .join(" · ");
  return (
    <CollectionRow
      title={objective.name}
      subtitle={subtitle}
      icon="flag-outline"
      accessibilityHint={`View ${getTermDisplay("storyTerm", { variant: "plural" })}`}
      onPress={() =>
        router.push(`/team/${objective.teamId}/objectives/${objective.id}`)
      }
    />
  );
};
