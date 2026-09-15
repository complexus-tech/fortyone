import { Redirect, useLocalSearchParams } from "expo-router";
import { legacyTeamStoriesHref } from "@/modules/teams/stories/team-story-navigation";

export default function ObjectiveDetailPage() {
  const params = useLocalSearchParams<{
    teamId: string;
    objectiveId: string;
  }>();
  return <Redirect href={legacyTeamStoriesHref(params, "objective")} />;
}
