import React from "react";
import { SafeContainer, Text } from "@/components/ui";
import { Header } from "./components/header";
import { useGlobalSearchParams } from "expo-router";
import { useTeamSprints } from "./hooks";

export const Sprints = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: sprints = [] } = useTeamSprints(teamId);
  return (
    <SafeContainer isFull>
      <Header />
      {sprints.map((sprint) => (
        <Text key={sprint.id}>{sprint.name}</Text>
      ))}
    </SafeContainer>
  );
};
