import { QueryState } from "@/components/ui/query-state";
import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { useLocalSearchParams } from "expo-router";
import { useTeamObjectives } from "./hooks";
import { List } from "./components/list";

export const Objectives = () => {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  const {
    data: objectives = [],
    isPending,
    error,
    refetch,
  } = useTeamObjectives(teamId);

  if (isPending) return <QueryState loading title="Loading objectives" />;
  if (error)
    return (
      <QueryState
        title="Could not load objectives"
        message={error.message}
        onRetry={() => {
          void refetch();
        }}
      />
    );

  return (
    <SafeContainer isFull>
      <Header />
      <List objectives={objectives} />
    </SafeContainer>
  );
};
