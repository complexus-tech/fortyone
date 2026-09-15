import { Redirect, useLocalSearchParams } from "expo-router";
import { legacyTeamStoriesHref } from "@/modules/teams/stories/team-story-navigation";

export default function SprintDetailPage() {
  const params = useLocalSearchParams<{ teamId: string; sprintId: string }>();
  return <Redirect href={legacyTeamStoriesHref(params, "sprint")} />;
}
