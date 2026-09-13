import { QueryState } from "@/components/ui/query-state";
import React from "react";
import { useLocalSearchParams } from "expo-router";
import { useTeamSprints } from "./hooks";
import { List } from "./components/list";

export const Sprints = () => {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  const {
    data: sprints = [],
    isPending,
    error,
    refetch,
  } = useTeamSprints(teamId);

  if (isPending) return <QueryState loading title="Loading sprints" />;
  if (error)
    return (
      <QueryState
        title="Could not load sprints"
        message={error.message}
        onRetry={() => {
          void refetch();
        }}
      />
    );

  return <List sprints={sprints} />;
};
