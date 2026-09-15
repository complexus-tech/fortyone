import type { Sprint } from "../types";
import { useRouter } from "expo-router";
import { useTerminology } from "@/hooks/use-terminology";
import { CollectionRow } from "@/modules/teams/stories/components/collection-row";
import {
  formatContextDates,
  getSprintTiming,
} from "@/modules/teams/stories/context-dates";

export const Card = ({ sprint }: { sprint: Sprint }) => {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  const context = getSprintTiming(sprint.startDate, sprint.endDate);
  const subtitle = [
    formatContextDates(sprint.startDate, sprint.endDate),
    context,
  ]
    .filter(Boolean)
    .join(" · ");
  return (
    <CollectionRow
      title={sprint.name}
      subtitle={subtitle}
      icon="play-circle-outline"
      accessibilityHint={`View ${getTermDisplay("storyTerm", { variant: "plural" })}`}
      onPress={() => router.push(`/team/${sprint.teamId}/sprints/${sprint.id}`)}
    />
  );
};
