import React from "react";
import { useGlobalSearchParams } from "expo-router";
import { useTeamSprints } from "./hooks";
import { List } from "./components/list";

export const Sprints = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: sprints = [] } = useTeamSprints(teamId);

  return <List sprints={sprints} />;
};
