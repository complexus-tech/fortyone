import React from "react";
import { useGlobalSearchParams } from "expo-router";
import { TeamStories } from "@/modules/teams/stories";

export default function TeamStoriesPage() {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();

  return <TeamStories teamId={teamId!} />;
}
