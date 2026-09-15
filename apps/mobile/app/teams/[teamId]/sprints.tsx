import { Redirect, useLocalSearchParams } from "expo-router";
import { legacyTeamStoriesHref } from "@/modules/teams/stories/team-story-navigation";

export default function SprintsScreen() {
  const params = useLocalSearchParams<{ teamId: string }>();
  return <Redirect href={legacyTeamStoriesHref(params, "sprint", true)} />;
}
