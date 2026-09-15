import { Redirect, useLocalSearchParams } from "expo-router";
import { legacyTeamStoriesHref } from "@/modules/teams/stories/team-story-navigation";

export default function ObjectivesScreen() {
  const params = useLocalSearchParams<{ teamId: string }>();
  return <Redirect href={legacyTeamStoriesHref(params, "objective", true)} />;
}
